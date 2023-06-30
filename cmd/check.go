/*
Copyright © 2023 Jacob Aronoff <jacob.aronoff@lightstep.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"

	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"github.com/lightstep/collector-cluster-check/pkg/checks/certmanager"
	"github.com/lightstep/collector-cluster-check/pkg/checks/dns"
	"github.com/lightstep/collector-cluster-check/pkg/checks/kubernetes"
	"github.com/lightstep/collector-cluster-check/pkg/checks/lightstep"
	"github.com/lightstep/collector-cluster-check/pkg/checks/oteloperator"
	"github.com/lightstep/collector-cluster-check/pkg/checks/prometheus"
	"github.com/lightstep/collector-cluster-check/pkg/dependencies"
)

type checkGroup struct {
	dependencies []dependencies.Initializer
	checkers     []checks.NewChecker
}

var (
	kubeConfig      string
	accessToken     string
	http            bool
	availableChecks = map[string]checkGroup{
		"metrics": {
			dependencies: []dependencies.Initializer{dependencies.MetricInitializer},
			checkers:     []checks.NewChecker{lightstep.NewMetricCheck},
		},
		"tracing": {
			dependencies: []dependencies.Initializer{dependencies.TraceInitializer},
			checkers:     []checks.NewChecker{lightstep.NewTraceCheck},
		},
		"preflight": {
			dependencies: []dependencies.Initializer{dependencies.KubernetesClientInitializer, dependencies.CustomResourceClientInitializer},
			checkers:     []checks.NewChecker{kubernetes.NewVersionCheck, prometheus.NewCheck, certmanager.NewCheck, oteloperator.NewCheck},
		},
		"dns": {
			dependencies: []dependencies.Initializer{},
			checkers:     []checks.NewChecker{dns.NewLookupCheck, dns.NewDialCheck},
		},
		"all": {
			dependencies: []dependencies.Initializer{dependencies.KubernetesClientInitializer, dependencies.CustomResourceClientInitializer, dependencies.MetricInitializer, dependencies.TraceInitializer},
			checkers:     []checks.NewChecker{dns.NewLookupCheck, dns.NewDialCheck, kubernetes.NewVersionCheck, prometheus.NewCheck, certmanager.NewCheck, oteloperator.NewCheck, lightstep.NewMetricCheck, lightstep.NewTraceCheck},
		},
	}
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [metrics|tracing|preflight|all]",
	Short: "runs one of multiple checks, use -h for more",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("must specify at least one check to run")
		}
		var validArgs []string
		for _, v := range cmd.ValidArgs {
			validArgs = append(validArgs, strings.Split(v, "\t")[0])
		}
		for _, v := range args {
			if _, ok := availableChecks[v]; !ok {
				return fmt.Errorf("invalid argument %q for %q", v, cmd.CommandPath())
			}
		}
		return nil
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var comps []string
		if len(args) == 0 {
			comps = cobra.AppendActiveHelp(comps, "You must choose at least one check to run")
		} else {
			for _, arg := range args {
				if _, ok := availableChecks[arg]; !ok {
					comps = cobra.AppendActiveHelp(comps, fmt.Sprintf("%s is not a valid check", arg))
				}
			}
		}
		return comps, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		for _, c := range args {
			group := availableChecks[c]
			var depResults []*checks.Check
			var results map[string]checks.CheckerResult
			var opts []checks.RunnerOption
			for _, d := range group.dependencies {
				runnerOption, checkResult := d.Apply(cmd.Context(), http, accessToken, kubeConfig)
				depResults = append(depResults, checkResult)
				if checkResult.IsFailure() {
					return
				}
				opts = append(opts, runnerOption)
			}
			runner := checks.NewRunner(group.checkers, opts...)
			results = runner.Run(cmd.Context())
			prettyPrintDependenciesResults(depResults)
			prettyPrint(results)
		}
	},
}

func prettyPrintDependenciesResults(results []*checks.Check) {
	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"dependency", "Result", "Message", "Error"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})
	for _, check := range results {
		prettyResult := "🟩"
		if !check.IsSuccess() {
			prettyResult = "🟥"
		}
		t.AppendRow(table.Row{check.Name, prettyResult, check.Message, check.Error}, rowConfigAutoMerge)
	}
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func prettyPrint(results map[string]checks.CheckerResult) {
	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"Checker", "Result", "Check Name", "Message", "Error"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})
	for checker, result := range results {
		for _, check := range result {
			prettyResult := "🟩"
			if !check.IsSuccess() {
				prettyResult = "🟥"
			}
			t.AppendRow(table.Row{checker, prettyResult, check.Name, check.Message, check.Error}, rowConfigAutoMerge)
		}
	}
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func init() {
	rootCmd.AddCommand(checkCmd)

	if home := homedir.HomeDir(); home != "" {
		checkCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeConfig", "", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		checkCmd.PersistentFlags().StringVarP(&kubeConfig, "kubeconfig", "", "", "absolute path to the kubeconfig file")
	}
	checkCmd.PersistentFlags().StringVarP(&accessToken, "accessToken", "", os.Getenv("LS_TOKEN"), "access token to send data to Lightstep")
	checkCmd.PersistentFlags().BoolVarP(&http, "http", "", false, "should telemetry be sent over http")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
