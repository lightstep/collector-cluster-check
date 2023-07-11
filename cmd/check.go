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
	"github.com/lightstep/collector-cluster-check/pkg/steps"
	"github.com/lightstep/collector-cluster-check/pkg/steps/dns"
	"github.com/lightstep/collector-cluster-check/pkg/steps/kubernetes"
	"github.com/lightstep/collector-cluster-check/pkg/steps/metrics"
	"github.com/lightstep/collector-cluster-check/pkg/steps/otel"
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

var (
	kubeConfig      string
	accessToken     string
	endpoint        string
	http            bool
	insecure        bool
	availableChecks = map[string]*steps.Check{
		"metrics": steps.NewCheck(
			"metrics",
			"Initializes a meter, creates a counter, flushes metrics",
			[]steps.Step{
				metrics.CreateCounter{},
				metrics.ShutdownMeter{},
			}),
		"tracing": steps.NewCheck(
			"tracing",
			"Initializes a trace provider, starts and finishes a trace, flushes the trace",
			[]steps.Step{
				traces.StartTrace{},
				traces.ShutdownTracer{},
			}),
		"preflight": steps.NewCheck(
			"preflight",
			"Runs preflight checks to ensure that a collector CRD can be created",
			[]steps.Step{
				kubernetes.Version{},
				kubernetes.NewCrdExists(steps.CertManagerCrdName),
				kubernetes.NewCrdExists(steps.OtelCrdName),
				kubernetes.NewCrdExists(steps.ServiceMonitorCrdName),
				kubernetes.NewPodRunning(steps.OtelOperatorSelector),
				kubernetes.NewPodRunning(steps.CertManagerSelector),
			}),
		"dns": steps.NewCheck(
			"dns",
			"Runs basic DNS checks to verify a connection to Lightstep from your local machine",
			[]steps.Step{
				dns.IPLookup{},
				dns.Ping{},
				dns.Dial{},
			}),
		"inflight": steps.NewCheck(
			"inflight",
			"Creates a collector, sends telemetry, queries that the telemetry was sent successfully to Lightstep",
			[]steps.Step{
				kubernetes.NewCrdExists(steps.OtelCrdName),
				otel.CreateCollector{},
				otel.PodWatcher{},
				kubernetes.StartPortForward{Port: 4317, LabelSelector: steps.LabelSelector},
				metrics.NewCreateCounter("localhost:4317", true),
				metrics.ShutdownMeter{},
				traces.NewStartTrace("localhost:4317", true),
				traces.ShutdownTracer{},
				kubernetes.FinishPortForward{Port: 4317, LabelSelector: steps.LabelSelector},
				kubernetes.StartPortForward{Port: 8888, LabelSelector: steps.LabelSelector},
				otel.QueryCollector{},
				kubernetes.FinishPortForward{Port: 8888, LabelSelector: steps.LabelSelector},
				otel.DeleteCollector{},
			}),
		"all": steps.NewCheck(
			"all",
			"Runs every available step",
			[]steps.Step{
				kubernetes.Version{},
				kubernetes.NewCrdExists(steps.CertManagerCrdName),
				kubernetes.NewCrdExists(steps.OtelCrdName),
				kubernetes.NewCrdExists(steps.ServiceMonitorCrdName),
				kubernetes.NewPodRunning(steps.OtelOperatorSelector),
				kubernetes.NewPodRunning(steps.CertManagerSelector),
				metrics.CreateCounter{},
				metrics.ShutdownMeter{},
				traces.StartTrace{},
				traces.ShutdownTracer{},
				dns.IPLookup{},
				dns.Ping{},
				dns.Dial{},
				otel.CreateCollector{},
				otel.PodWatcher{},
				kubernetes.StartPortForward{Port: 4317, LabelSelector: steps.LabelSelector},
				metrics.NewCreateCounter("localhost:4317", true),
				metrics.ShutdownMeter{},
				traces.NewStartTrace("localhost:4317", true),
				traces.ShutdownTracer{},
				kubernetes.FinishPortForward{Port: 4317, LabelSelector: steps.LabelSelector},
				kubernetes.StartPortForward{Port: 8888, LabelSelector: steps.LabelSelector},
				otel.QueryCollector{},
				kubernetes.FinishPortForward{Port: 8888, LabelSelector: steps.LabelSelector},
				otel.DeleteCollector{},
			}),
	}
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check [metrics|tracing|preflight|inflight|all]",
	Short: "Check can run one of multiple checks, use -h for more",
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
			conf := GetConfig()
			deps := steps.NewDependencies()
			depResults, checkResults := group.Run(cmd.Context(), deps, conf)
			prettyPrintDependenciesResults(depResults)
			prettyPrint(checkResults)
		}
	},
}

func prettyPrintDependenciesResults(checkResults []steps.Results) {
	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"dependency", "Result", "Message", "Error"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})
	for _, results := range checkResults {
		for _, result := range results.Steps() {
			prettyResult := "🟩"
			if !result.Successful() {
				prettyResult = "🟥"
			}
			t.AppendRow(table.Row{results.StepName(), prettyResult, result.Message(), result.Err()}, rowConfigAutoMerge)
		}
	}
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func prettyPrint(checkResults []steps.Results) {
	t := table.NewWriter()
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t.AppendHeader(table.Row{"Checker", "Result", "Message", "Error"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
	})
	for _, results := range checkResults {
		for _, result := range results.Steps() {
			prettyResult := "🟩"
			if !result.Successful() {
				prettyResult = "🟥"
			}
			t.AppendRow(table.Row{results.StepName(), prettyResult, result.Message(), result.Err()}, rowConfigAutoMerge)
		}
	}
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.Style().Options.SeparateRows = true
	t.Render()
}

func GetConfig() *steps.Config {
	return &steps.Config{
		Endpoint:   endpoint,
		Insecure:   insecure,
		Http:       http,
		Token:      accessToken,
		KubeConfig: kubeConfig,
	}
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
	checkCmd.SetHelpFunc(func(command *cobra.Command, i []string) {
		// If help was called only on the base command
		command.Println(checkCmd.UsageString())
		if len(i) == 2 {
			command.Println(command.Short)
			return
		} else if len(i) == 3 {
			// We expect the check name to be at the first index
			name := i[1]
			check, ok := availableChecks[name]
			if !ok {
				command.Println(fmt.Sprintf("command %s not found", name))
				command.Println(command.Short)
				return
			}
			command.Println(fmt.Sprintf("----%s check description----", check.Name()))
			command.Println(check.Description())
		}
	})

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
