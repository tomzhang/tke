/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the “License”); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an “AS IS” BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"fmt"
	"reflect"
	"time"
	v1 "tkestack.io/tke/api/auth/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	clientset "tkestack.io/tke/api/client/clientset/versioned"
	authv1informer "tkestack.io/tke/api/client/informers/externalversions/auth/v1"
	authv1lister "tkestack.io/tke/api/client/listers/auth/v1"
	"tkestack.io/tke/pkg/business/controller/project/deletion"
	controllerutil "tkestack.io/tke/pkg/controller"
	"tkestack.io/tke/pkg/util/log"
	"tkestack.io/tke/pkg/util/metrics"
)

const (
	// policyDeletionGracePeriod is the time period to wait before processing a received channel event.
	// This allows time for the following to occur:
	// * lifecycle admission plugins on HA apiservers to also observe a channel
	//   deletion and prevent new objects from being created in the terminating channel
	// * non-leader etcd servers to observe last-minute object creations in a channel
	//   so this controller's cleanup can actually clean up all objects
	policyDeletionGracePeriod = 5 * time.Second

	controllerName = "policy-controller"
)

// Controller is responsible for performing actions dependent upon a project phase.
type Controller struct {
	client       clientset.Interface
	cache        *policyCache
	queue        workqueue.RateLimitingInterface
	lister       authv1lister.PolicyLister
	listerSynced cache.InformerSynced
	// helper to delete all resources in the project when the project is deleted.
	projectedResourcesDeleter deletion.ProjectedResourcesDeleterInterface
}

// NewController creates a new Project object.
func NewController(client clientset.Interface, policyInformer authv1informer.APIKeyInformer, resyncPeriod time.Duration) *Controller {
	// create the controller so we can inject the enqueue function
	controller := &Controller{
		client:                    client,
		cache:                     &policyCache{policyMap: make(map[string]*cachedPolicy)},
		queue:                     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), controllerName),
	}

	if client != nil && client.AuthV1().RESTClient().GetRateLimiter() != nil {
		_ = metrics.RegisterMetricAndTrackRateLimiterUsage("policy_controller", client.AuthV1().RESTClient().GetRateLimiter())
	}

	policyInformer.Informer().AddEventHandlerWithResyncPeriod(
		cache.ResourceEventHandlerFuncs{
			//AddFunc: controller.enqueue,
			UpdateFunc: func(oldObj, newObj interface{}) {
				old, ok1 := oldObj.(*v1.Policy)
				cur, ok2 := newObj.(*v1.Policy)
				if ok1 && ok2 && controller.needsUpdate(old, cur) {
					log.Info("Update enqueue")
					controller.enqueue(newObj)
				}
			},
			DeleteFunc: controller.enqueue,
		},
		resyncPeriod,
	)
	controller.lister = policyInformer.Lister()
	controller.listerSynced = policyInformer.Informer().HasSynced
	return controller
}

// obj could be an *v1.Project, or a DeletionFinalStateUnknown marker item.
func (c *Controller) enqueue(obj interface{}) {
	key, err := controllerutil.KeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}
	c.queue.AddAfter(key, policyDeletionGracePeriod)
}

func (c *Controller) needsUpdate(old *v1.Policy, new *v1.Policy) bool {
	if old.UID != new.UID {
		return true
	}

	if !reflect.DeepEqual(old.Spec, new.Spec) {
		return true
	}

	return false
}


// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers.
func (c *Controller) Run(workers int, stopCh <-chan struct{}) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	log.Info("Starting project controller")
	defer log.Info("Shutting down project controller")

	if ok := cache.WaitForCacheSync(stopCh, c.listerSynced); !ok {
		log.Error("Failed to wait for project caches to sync")
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.worker, time.Second, stopCh)
	}

	<-stopCh
}

// worker processes the queue of project objects.
// Each project can be in the queue at most once.
// The system ensures that no two workers can process
// the same project at the same time.
func (c *Controller) worker() {
	workFunc := func() bool {
		key, quit := c.queue.Get()
		if quit {
			return true
		}
		defer c.queue.Done(key)

		requeue, err := c.syncItem(key.(string))
		if err == nil && !requeue {
			// no error, forget this entry and return
			c.queue.Forget(key)
			return false
		}

		// rather than wait for a full resync, re-add the project to the queue to be processed
		c.queue.AddRateLimited(key)
		runtime.HandleError(err)
		return false
	}

	for {
		quit := workFunc()

		if quit {
			return
		}
	}
}

// syncItem will sync the ApiKey with the given key if it has had
// its expectations fulfilled, meaning the apikey has been deleted by user but not expired.
func (c *Controller) syncItem(key string) (bool, error) {
	startTime := time.Now()

	defer func() {
		log.Info("Finished syncing policy", log.String("apikey", key), log.Duration("processTime", time.Since(startTime)))
	}()

	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return false, err
	}

	policy, err := c.lister.Get(name)
	switch {
	case errors.IsNotFound(err):
		log.Infof("Api key has been deleted %v", key)
		return false, nil
	case err != nil:
		log.Warn("Unable to retrieve policy from store", log.String("policy name", key), log.Err(err))
	default:
		// api key has been deleted check whether it has expired
		log.Info("Create policy", log.Any("policy", policy))
	}
	return false, nil
}