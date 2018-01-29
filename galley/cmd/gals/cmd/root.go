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
	"flag"
	"fmt"

	"github.com/spf13/cobra"
	"istio.io/istio/pkg/log"

	"istio.io/istio/galley/cmd/shared"
	"istio.io/istio/pkg/version"
)

type rootArgs struct {
	// kubeconfigPath is the path to the kube conf file. If not specified, then in-cluster config will be
	// used.
	kubeconfigPath string

	loggingOptions *log.Options
}

// GetRootCmd returns the root of the cobra command-tree.
func GetRootCmd(args []string, printf, fatalf shared.FormatFn) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "galc",
		Short: "Utility to interact with Galley.",
		Long: "This command lets you interact with a running instance of\n" +
			"Galley. Note that you need a pretty good understanding of Galley\n" +
			"in order to use this command.",

		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("'%s' is an invalid argument", args[0])
			}
			return nil
		},
	}
	rootCmd.SetArgs(args)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	// hack to make flag.Parsed return true such that glog is happy
	// about the flags having been parsed
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	/* #nosec */
	_ = fs.Parse([]string{})
	flag.CommandLine = fs

	ra := &rootArgs{
		loggingOptions: log.NewOptions(),
	}

	rootCmd.PersistentFlags().StringVarP(&ra.kubeconfigPath, "kubeconfig", "c", "",
		"kube config file path")

	rootCmd.AddCommand(initCmd(ra, fatalf))
	rootCmd.AddCommand(typesCmd(ra, printf, fatalf))
	rootCmd.AddCommand(copyCmd(ra, fatalf))
	rootCmd.AddCommand(deleteCmd(ra, fatalf))
	rootCmd.AddCommand(serverCmd(ra, printf, fatalf))
	rootCmd.AddCommand(version.CobraCommand())

	ra.loggingOptions.AttachCobraFlags(rootCmd)

	return rootCmd
}
