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
	"istio.io/istio/galley/pkg/config/sync"
	"istio.io/istio/pkg/log"

	//"istio.io/istio/galley/pkg/config/kube"

	"istio.io/istio/galley/cmd/shared"
)

func serverCmd(ra *rootArgs, printf, fatalf shared.FormatFn) *cobra.Command {

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Starts Galley as a server",
		Run: func(cmd *cobra.Command, args []string) {
			runServer(ra, printf, fatalf)
		},
	}

	return serverCmd
}

func runServer(ra *rootArgs, printf, fatalf shared.FormatFn) (err error) {
	printf("Galley started with\n%s", ra)

	if err = log.Configure(ra.loggingOptions); err != nil {
		return
	}

	var controller *sync.Controller
	if controller, err = sync.NewController(ra.kubeconfigPath); err != nil {
		return
	}

	var done chan struct{}
	if done, err = controller.Sync(); err != nil {
		return
	}

	<-done
	return
}
