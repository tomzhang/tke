/*
 * Tencent is pleased to support the open source community by making TKEStack
 * available.
 *
 * Copyright (C) 2012-2019 Tencent. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use
 * this file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/Apache-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OF ANY KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations under the License.
 */

// Code generated by client-gen. DO NOT EDIT.

package internalversion

import (
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	auth "tkestack.io/tke/api/auth"
	scheme "tkestack.io/tke/api/client/clientset/internalversion/scheme"
)

// RolesGetter has a method to return a RoleInterface.
// A group's client should implement this interface.
type RolesGetter interface {
	Roles() RoleInterface
}

// RoleInterface has methods to work with Role resources.
type RoleInterface interface {
	Create(*auth.Role) (*auth.Role, error)
	Update(*auth.Role) (*auth.Role, error)
	UpdateStatus(*auth.Role) (*auth.Role, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*auth.Role, error)
	List(opts v1.ListOptions) (*auth.RoleList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *auth.Role, err error)
	RoleExpansion
}

// roles implements RoleInterface
type roles struct {
	client rest.Interface
}

// newRoles returns a Roles
func newRoles(c *AuthClient) *roles {
	return &roles{
		client: c.RESTClient(),
	}
}

// Get takes name of the role, and returns the corresponding role object, and an error if there is any.
func (c *roles) Get(name string, options v1.GetOptions) (result *auth.Role, err error) {
	result = &auth.Role{}
	err = c.client.Get().
		Resource("roles").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Roles that match those selectors.
func (c *roles) List(opts v1.ListOptions) (result *auth.RoleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &auth.RoleList{}
	err = c.client.Get().
		Resource("roles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested roles.
func (c *roles) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("roles").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a role and creates it.  Returns the server's representation of the role, and an error, if there is any.
func (c *roles) Create(role *auth.Role) (result *auth.Role, err error) {
	result = &auth.Role{}
	err = c.client.Post().
		Resource("roles").
		Body(role).
		Do().
		Into(result)
	return
}

// Update takes the representation of a role and updates it. Returns the server's representation of the role, and an error, if there is any.
func (c *roles) Update(role *auth.Role) (result *auth.Role, err error) {
	result = &auth.Role{}
	err = c.client.Put().
		Resource("roles").
		Name(role.Name).
		Body(role).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *roles) UpdateStatus(role *auth.Role) (result *auth.Role, err error) {
	result = &auth.Role{}
	err = c.client.Put().
		Resource("roles").
		Name(role.Name).
		SubResource("status").
		Body(role).
		Do().
		Into(result)
	return
}

// Delete takes name of the role and deletes it. Returns an error if one occurs.
func (c *roles) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("roles").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *roles) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("roles").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched role.
func (c *roles) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *auth.Role, err error) {
	result = &auth.Role{}
	err = c.client.Patch(pt).
		Resource("roles").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
