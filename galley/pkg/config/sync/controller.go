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

package sync

import (
	"istio.io/istio/galley/pkg/config/model/v1"
	"istio.io/istio/galley/pkg/config/sync/crd"
	"istio.io/istio/galley/pkg/config/sync/resource"
	"istio.io/istio/pkg/log"
	ext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	//	// import GKE cluster authentication plugin. Otherwise we get `No Auth Provider found for name "gcp"`
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Controller struct {
	crd      *crd.Client
	internal *resource.Client
	public   *resource.Client

	synchronizers []*resource.Synchronizer
}

func NewController(kubeconfigPath string) (controller *Controller, err error) {
	var config *rest.Config
	if config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath); err != nil {
		return
	}

	c := &Controller{}
	if c.crd, err = crd.NewClient(config); err != nil {
		return
	}

	var defs []ext.CustomResourceDefinition
	if defs, err = c.crd.GetAll(v1.PublicGroupVersion()); err != nil {
		return
	}

	scheme := runtime.NewScheme()
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	addTypesToScheme(scheme, defs)

	if c.public, err = resource.NewClient(config, v1.PublicGroupVersion(), scheme); err != nil {
		return
	}

	if defs, err = c.crd.GetAll(v1.InternalGroupVersion()); err != nil {
		return
	}

	scheme = runtime.NewScheme()
	metav1.AddToGroupVersion(scheme, schema.GroupVersion{Version: "v1"})
	addTypesToScheme(scheme, defs)

	if c.internal, err = resource.NewClient(config, v1.InternalGroupVersion(), scheme); err != nil {
		return
	}

	controller = c
	return
}

func (c *Controller) Crds() *crd.Client {
	return c.crd
}

func (c *Controller) CopyCRDs() error {
	// Retrieve the definitions for the core Istio CRDs.
	definitions, err := c.crd.GetAll(v1.PublicGroupVersion())
	if err != nil {
		return err
	}

	definitions = rewriteCrds(v1.InternalGroupVersion(), definitions)
	return c.crd.Upsert(definitions)
}

func rewriteCrds(gv schema.GroupVersion, originals []ext.CustomResourceDefinition) []ext.CustomResourceDefinition {
	result := make([]ext.CustomResourceDefinition, len(originals))

	for i, d := range originals {
		d.Spec.Group = gv.Group
		d.Spec.Version = gv.Version
		d.Name = d.Spec.Names.Plural + "." + d.Spec.Group
		d.ResourceVersion = ""
		delete(d.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
		result[i] = d
	}

	return result
}
func (c *Controller) CopyResources() error {
	log.Debugf("Syncing 'ac%v' => '%v'", v1.PublicGroupVersion(), v1.InternalGroupVersion())
	defs, err := c.crd.GetAll(v1.PublicGroupVersion())
	if err != nil {
		return err
	}
	// Collect resourceTypes of types
	resourceTypes := make([]string, 0, len(defs))
	for _, d := range defs {
		log.Debugf("Found resource type to sync: '%s'", d.Spec.Names.Plural)
		resourceTypes = append(resourceTypes, d.Spec.Names.Plural)
	}

	for _, resourceType := range resourceTypes {
		log.Debugf("Syncing resource type: '%s'", resourceType)
		resourceList, err := c.public.GetAll(resourceType)
		if err != nil {
			log.Debugf("Error retrieving resource: '%v'", err)
			return err
		}

		gv := v1.InternalGroupVersion()

		// TODO: delete extra ones as well.
		for _, r := range resourceList.Items {
			r.SetUID("")
			r.SetAPIVersion("")
			r.SetSelfLink("")
			r.SetResourceVersion("")
			r.SetGenerateName("")
			r.SetAPIVersion(gv.String())
			ann := r.GetAnnotations()
			delete(ann, "kubectl.kubernetes.io/last-applied-configuration")
			r.SetAnnotations(ann)

			log.Debugf("  Syncing resource: %s(%s)", r.GetName(), r.GetKind())

			err := c.internal.Upsert(resourceType, &r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Controller) DeleteResources() error {
	// TODO: ACtually delete resources as well
	return c.crd.DeleteAll(v1.InternalGroupVersion())
}

func (c *Controller) Sync() (chan struct{}, error) {

	defs, err := c.crd.GetAll(v1.PublicGroupVersion())
	if err != nil {
		return nil, err
	}

	c.synchronizers = make([]*resource.Synchronizer, len(defs))
	for i, def := range defs {
		c.synchronizers[i] = resource.NewSynchronizer(def, c.public, c.internal)
	}

	for _, s := range c.synchronizers {
		err = s.Start()
		if err != nil {
			break
		}
	}

	if err != nil {
		for _, s := range c.synchronizers {
			s.Close()
		}

		return nil, err
	}

	// TODO: Use the done channel to close
	done := make(chan struct{}, 1)

	return done, nil
}
