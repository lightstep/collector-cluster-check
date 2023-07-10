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
	"github.com/lightstep/collector-cluster-check/pkg/checks"
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/certmanager"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dns"
	"github.com/lightstep/collector-cluster-check/pkg/steps/kubernetes"
	"github.com/lightstep/collector-cluster-check/pkg/steps/metrics"
	"github.com/lightstep/collector-cluster-check/pkg/steps/otel"
	"github.com/lightstep/collector-cluster-check/pkg/steps/prometheus"
	"github.com/lightstep/collector-cluster-check/pkg/steps/traces"
	"os"
	"path/filepath"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	//"github.com/lightstep/collector-cluster-check/pkg/checks"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/certmanager"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/dns"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/kubernetes"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/lightstep"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/oteloperator"
	//"github.com/lightstep/collector-cluster-check/pkg/checks/prometheus"
	//"github.com/lightstep/collector-cluster-check/pkg/dependencies"
)

type checkGroup struct {
	steps []steps.Step
}

var (
	kubeConfig      string
	accessToken     string
	endpoint        string
	http            bool
	insecure        bool
	availableChecks = map[string]checkGroup{
		"metrics": {
			steps: []steps.Step{
				metrics.CreateCounter{},
				metrics.ShutdownMeter{},
			},
		},
		"tracing": {
			steps: []steps.Step{
				traces.StartTrace{},
				traces.ShutdownTracer{},
			},
		},
		"preflight": {
			steps: []steps.Step{
				kubernetes.Version{},
				prometheus.CrdExists{},
				certmanager.CrdExists{},
				otel.CrdExists{},
			},
		},
		"dns": {
			steps: []steps.Step{
				dns.IPLookup{},
				dns.Ping{},
				dns.Dial{},
			},
		},
		"inflight": {
			steps: []steps.Step{
				otel.CrdExists{},
				otel.CreateCollector{},
				otel.PodWatcher{},
				otel.PortForward{Port: 4317},
				metrics.CreateCounter{},
				metrics.ShutdownMeter{},
				traces.StartTrace{},
				traces.ShutdownTracer{},
				otel.DeleteCollector{},
			},
		},
		//"dns": {
		//	dependencies: []dependencies.Initializer{},
		//	checkers:     []checks.NewChecker{dns.NewLookupCheck, dns.NewDialCheck},
		//},
		//"inflight": {
		//	dependencies: []dependencies.Initializer{dependencies.DynamicClientInitializer, dependencies.KubernetesClientInitializer, dependencies.KubeConfigInitializer, dependencies.OtelCollectorConfigInitializer, dependencies.OtelColMetricInitializer, dependencies.OtelColTraceInitializer},
		//	checkers:     []checks.NewChecker{runcrd.NewRunCollectorCheck},
		//},
		//"all": {
		//	dependencies: []dependencies.Initializer{dependencies.KubernetesClientInitializer, dependencies.CustomResourceClientInitializer, dependencies.MetricInitializer, dependencies.TraceInitializer},
		//	checkers:     []checks.NewChecker{dns.NewLookupCheck, dns.NewDialCheck, kubernetes.NewVersionCheck, prometheus.NewCheck, certmanager.NewCheck, oteloperator.NewCheck, lightstep.NewMetricCheck, lightstep.NewTraceCheck},
		//},
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
			results := map[string]steps.Result{}
			conf := getConfig()
			deps := steps.NewDependencies()
			for _, step := range group.steps {
				for _, dep := range step.Dependencies(conf) {
					opt, r := dep.Run(cmd.Context(), deps)
					results[dep.Name()] = r
					if !r.Successful() && !r.ShouldContinue() {
						break
					}
					opt(deps)
				}
				opt, r := step.Run(cmd.Context(), deps)
				results[step.Name()] = r
				if !r.Successful() && !r.ShouldContinue() {
					break
				}
				opt(deps)
			}
			//prettyPrintDependenciesResults(depResults)
			prettyPrint(results)
		}
	},
}

func getConfig() *steps.Config {
	return &steps.Config{
		Endpoint:   endpoint,
		Insecure:   insecure,
		Http:       http,
		Token:      accessToken,
		KubeConfig: kubeConfig,
	}
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

func prettyPrint(results map[string]steps.Result) {
	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"Checker", "Result", "Message", "Error"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})
	for checker, result := range results {
		prettyResult := "🟩"
		if !result.Successful() {
			prettyResult = "🟥"
		}
		t.AppendRow(table.Row{checker, prettyResult, result.Message(), result.Err()}, rowConfigAutoMerge)
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
	checkCmd.PersistentFlags().StringVarP(&endpoint, "endpoint", "", "ingest.lightstep.com:443", "destination for OTLP data")
	checkCmd.PersistentFlags().BoolVarP(&http, "http", "", false, "should telemetry be sent over http")
	checkCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "", false, "should telemetry be sent insecurely")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
