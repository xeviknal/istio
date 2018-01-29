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
	"fmt"

	"istio.io/istio/galley/pkg/config/model/v1"
	"istio.io/istio/pkg/log"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"

	//	// import GKE cluster authentication plugin. Otherwise we get `No Auth Provider found for name "gcp"`
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

type Synchronizer struct {
	friendlyName string
	source       *Client
	destination  *Client
	definition   v1beta1.CustomResourceDefinition
	shutdown     chan struct{}
	done         chan struct{}
	watch        watch.Interface
}

func NewSynchronizer(definition v1beta1.CustomResourceDefinition, source *Client, destination *Client) (s *Synchronizer) {
	friendlyName := fmt.Sprintf("%s/%s/%s", definition.Spec.Group, definition.Spec.Version, definition.Spec.Names.Plural)
	return &Synchronizer{
		friendlyName: friendlyName,
		source:       source,
		destination:  destination,
		definition:   definition,
		shutdown:     make(chan struct{}, 1),
		done:         make(chan struct{}, 1),
	}
}

func (s *Synchronizer) Start() (err error) {

	log.Infof("Starting watch of '%s'", s.friendlyName)
	var w watch.Interface
	if w, err = s.source.Watch(s.definition.Spec.Names.Plural, ""); err != nil {
		log.Debugf("Error while watch start: '%s', err:='%v'", s.friendlyName, err)
		return
	}

	go func() {
		done := false
		currentRV := ""
		for !done {
			skip := false
			select {
			case event := <-w.ResultChan():
				var u *unstructured.Unstructured
				if event.Object != nil {
					u = event.Object.(*unstructured.Unstructured)
					if event.Type != watch.Deleted && u.GetResourceVersion() == currentRV {
						skip = true
					} else {
						currentRV = u.GetResourceVersion()
						log.Debugf("Incoming event: %v => %s/%s @%s", event.Type, u.GetNamespace(), u.GetName(), u.GetResourceVersion())
					}
				}

				if skip {
					log.Debugf("Skipping re-read object")
					continue
				}

				switch event.Type {
				case watch.Added, watch.Modified:
					ru := rewriteResource(s.definition.Spec.Names.Kind, u)
					err := s.destination.Upsert(s.definition.Spec.Names.Plural, ru)
					if err != nil {
						log.Errorf("Error during synchronization (upsert): %v", err)
					}

				case watch.Deleted:
					err := s.destination.Delete(s.definition.Spec.Names.Plural, u.GetNamespace(), u.GetName())
					if err != nil {
						log.Errorf("Error during synchronization (delete): %v", err)
					}
				case watch.Error:
					log.Errorf("Unknown error encountered: %v", err)
					s.done <- struct{}{}
					done = true

				case "":
					// TODO: This seems to be some sort of timeout issue. re-init things again.
					log.Debugf("Reinitializing watch of '%s'", s.friendlyName)
					w.Stop()
					// TODO: Pass the right ResourceVersion so we don't end up re-reading the whole world again.
					if w, err = s.source.Watch(s.definition.Spec.Names.Plural, currentRV); err != nil {
						log.Errorf("Error while watch (re)start: '%s', err:='%v'", s.friendlyName, err)
						done = true
					}

				default:
					log.Errorf("Unknown event encountered: %v", event.Type)
					done = true
				}

			case <-s.shutdown:
				done = true
			}
		}

		log.Infof("Stopping watch on: '%s'", s.friendlyName)
		w.Stop()
		s.done <- struct{}{}
		return
	}()

	return nil
}

func (s *Synchronizer) Close() {
	s.shutdown <- struct{}{}
	<-s.done
}

func rewriteResource(resource string, original *unstructured.Unstructured) *unstructured.Unstructured {
	result := original.DeepCopy()
	result.SetAPIVersion(v1.InternalGroupVersion().String())
	result.SetResourceVersion("")
	result.SetUID("")
	result.SetSelfLink("")
	result.SetGenerateName("")
	result.SetKind(resource)
	result.SetGroupVersionKind(v1.InternalGroupVersion().WithKind(resource))

	return result
}
