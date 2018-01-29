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

package cmd

import (
	"github.com/spf13/cobra"
	"istio.io/istio/galley/cmd/shared"
	"istio.io/istio/galley/pkg/config/model/v1"
	"istio.io/istio/galley/pkg/config/sync"
	"istio.io/istio/pkg/log"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func typesCmd(rootArgs *rootArgs, printf, fatalf shared.FormatFn) *cobra.Command {
	return &cobra.Command{
		Use:   "types",
		Short: "Lists public and internal Istio configuration CRDs.",
		Long:  "Lists the current public and internal configuration CRDs that are known by the API server.",

		Run: func(cmd *cobra.Command, args []string) {
			if err := listTypes(rootArgs, printf, fatalf); err != nil {
				fatalf("%v", err)
			}
		}}
}

func listTypes(rootArgs *rootArgs, printf, fatalf shared.FormatFn) (err error) {
	if err = log.Configure(rootArgs.loggingOptions); err != nil {
		return
	}

	var controller *sync.Controller
	if controller, err = sync.NewController(rootArgs.kubeconfigPath); err != nil {
		return
	}

	if err = print(printf, controller, v1.PublicGroupVersion()); err != nil {
		return
	}

	printf("\n")

	err = print(printf, controller, v1.InternalGroupVersion())
	return
}

func print(printf shared.FormatFn, controller *sync.Controller, version schema.GroupVersion) (err error) {

	var definitions []v1beta1.CustomResourceDefinition
	if definitions, err = controller.Crds().GetAll(version); err != nil {
		return
	}

	printf("Custom Resource Definitions for '%s':", version.String())
	for _, d := range definitions {
		printf(" %s", d.Name)
	}

	return
}
