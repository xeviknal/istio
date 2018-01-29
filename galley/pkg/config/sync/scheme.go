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
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func addTypesToScheme(s *runtime.Scheme, definitions []v1beta1.CustomResourceDefinition) (*runtime.Scheme, error) {
	builder := runtime.NewSchemeBuilder(func(s *runtime.Scheme) error {
		for _, defn := range definitions {
			// Add the object itself
			gvk := schema.GroupVersionKind{
				Group:   defn.Spec.Group,
				Version: defn.Spec.Version,
				Kind:    defn.Spec.Names.Kind,
			}
			o := &unstructured.Unstructured{}
			o.SetAPIVersion(defn.Spec.Version)
			o.SetKind(defn.Spec.Names.Kind)
			s.AddKnownTypeWithName(gvk, o)

			// Add the collection object.
			gvk = schema.GroupVersionKind{
				Group:   defn.Spec.Group,
				Version: defn.Spec.Version,
				Kind:    defn.Spec.Names.ListKind,
			}
			c := &unstructured.UnstructuredList{}
			o.SetAPIVersion(defn.Spec.Version)
			o.SetKind(defn.Spec.Names.ListKind)
			s.AddKnownTypeWithName(gvk, c)
		}
		return nil
	})

	err := builder.AddToScheme(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
