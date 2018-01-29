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

package resource

import (
	"istio.io/istio/pkg/log"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/rest"
)

type Client struct {
	client *rest.RESTClient
}

func NewClient(cfg *rest.Config, gv schema.GroupVersion, scheme *runtime.Scheme) (*Client, error) {
	configShallowCopy := *cfg

	configShallowCopy.GroupVersion = &gv
	configShallowCopy.APIPath = "/apis"
	configShallowCopy.ContentType = runtime.ContentTypeJSON

	if scheme == nil {
		scheme := runtime.NewScheme()
		metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	}

	configShallowCopy.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	client, err := rest.RESTClientFor(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func (a *Client) GetAll(resource string) (*unstructured.UnstructuredList, error) {
	var list unstructured.UnstructuredList
	err := a.client.Get().
		Resource(resource).
		Do().
		Into(&list)

	return &list, err
}

func (a *Client) Upsert(resource string, o *unstructured.Unstructured) error {
	var existing unstructured.Unstructured
	err := a.client.Get().
		Namespace(o.GetNamespace()).
		Resource(resource).
		Name(o.GetName()).
		Do().Into(&existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		log.Debugf("Inserting [%s] %s/%s", o.GroupVersionKind(), o.GetNamespace(), o.GetName())
		return a.client.Post().
			Namespace(o.GetNamespace()).
			Resource(resource).
			Body(o).
			Do().
			Error()
	}

	log.Debugf("Updating [%s] %s/%s", o.GroupVersionKind(), o.GetNamespace(), o.GetName())
	o.SetResourceVersion(existing.GetResourceVersion())

	return a.client.Put().
		Namespace(o.GetNamespace()).
		Resource(resource).
		Name(o.GetName()).
		Body(o).
		Do().
		Error()
}

func (a *Client) Watch(resource string, resourceVersion string) (watch.Interface, error) {
	opts := metav1.ListOptions{
		Watch:           true,
		ResourceVersion: resourceVersion,
	}

	return a.client.Get().
		Resource(resource).
		VersionedParams(&opts, metav1.ParameterCodec).
		Watch()
}

func (a *Client) Delete(resource string, namespace string, name string) error {
	log.Debugf("Deleting [%s] %s/%s", resource, namespace, name)
	return a.client.Delete().
		Resource(resource).
		Namespace(namespace).
		Name(name).
		Do().
		Error()
}
