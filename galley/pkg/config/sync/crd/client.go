// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package crd

import (
	"istio.io/istio/pkg/log"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"

	ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client that implements higher-level functionality for accessing CRDs.
type Client struct {
	iface v1beta1.CustomResourceDefinitionInterface
}

// NewClient returns a new crd.Client
func NewClient(config *rest.Config) (client *Client, err error) {
	var set *clientset.Clientset
	if set, err = clientset.NewForConfig(config); err != nil {
		return
	}

	return &Client{
		iface: set.ApiextensionsV1beta1().CustomResourceDefinitions(),
	}, nil
}

// Get returns CRD definitions for the supplied group/version
func (c *Client) GetAll(gv schema.GroupVersion) ([]ext.CustomResourceDefinition, error) {
	var result []ext.CustomResourceDefinition

	cont := ""
	for {
		opts := metav1.ListOptions{
			Continue: cont,
		}

		list, err := c.iface.List(opts)
		if err != nil {
			return nil, err
		}

		for _, item := range list.Items {
			// TODO(ozben): Can we use a field selector for this?
			if item.Spec.Group == gv.Group && item.Spec.Version == gv.Version {
				result = append(result, item)
			}
		}

		cont = list.Continue
		if cont == "" {
			break
		}
	}

	return result, nil
}

// Upsert the given custom resource definitions
func (c *Client) Upsert(resources []ext.CustomResourceDefinition) error {

	for _, r := range resources {
		_, err := c.iface.Create(&r)
		if err != nil && apierrors.IsAlreadyExists(err) {
			log.Debugf("Resource already exists: %v", err)
			var existing *ext.CustomResourceDefinition
			existing, err = c.iface.Get(r.Name, metav1.GetOptions{})
			if err != nil {
				r.UID = existing.UID
				_, err = c.iface.Update(&r)
			}
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteAll crds that are in the given group, version.
func (c *Client) DeleteAll(gv schema.GroupVersion) error {
	// TODO: Can we use field selector and do a DeleteCollection?

	resources, err := c.GetAll(gv)
	if err != nil {
		return err
	}

	do := metav1.DeleteOptions{
	// TODO
	}
	for _, r := range resources {
		err = c.iface.Delete(r.Name, &do)
		if err != nil {
			return err
		}
	}

	return nil
}
