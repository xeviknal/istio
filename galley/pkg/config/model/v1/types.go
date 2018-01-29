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

package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PublicGroupVersion returns the schema.GroupVersion object for Istio resources.
func PublicGroupVersion() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   PublicAPIGroup,
		Version: APIVersion,
	}
}

// InternalGroupVersion returns the schema.GroupVersion object for Galley resources.
func InternalGroupVersion() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   InternalAPIGroup,
		Version: APIVersion,
	}
}

// TODO: Add a list of minimally required set of CRDs
