# collector-cluster-check

This tool allows you to test various parts of your infrastructure and observability stack.
The goal is to provide you with confidence that you are able to effectively use the OpenTelemetry Operator,
create OpenTelemetry Collectors, and send data to your desired destination.

```
Usage:
collector-cluster-check [command]

Available Commands:
check       runs one of multiple checks, use -h for more
completion  Generate the autocompletion script for the specified shell
help        Help about any command

Flags:
--config string   config file (default is $HOME/.collector-cluster-check.yaml)
-h, --help            help for collector-cluster-check
-t, --toggle          Help message for toggle

Use "collector-cluster-check [command] --help" for more information about a command.

runs one of multiple checks, use -h for more
```

## `check` Command

```
Usage:
  collector-cluster-check check [metrics|tracing|preflight|all] [flags]

Flags:
      --accessToken string   access token to send data to Lightstep
  -h, --help                 help for check
      --http                 should telemetry be sent over http
      --kubeConfig string    (optional) absolute path to the kubeconfig file (default "/Users/jacob.aronoff/.kube/config")

Global Flags:
      --config string   config file (default is $HOME/.collector-cluster-check.yaml)
```