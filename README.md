# infrastructure-monitor

Monitoring the state of application clusters

## Structure

On each application cluster, `infrastructure-monitor-slave` listens to Kubernetes events and translates
them into Platform-relevant metrics. These are exposed on `:8424/metrics` for Prometheus to scrape.

`infrastructure-monitor-slave` also exposes a gRPC API on port 8422 that provides a cluster resource
summary and platform statistics by querying Prometheus. It also allows for generic queries.

Ultimately, `public-api` provides an interface to retrieve the cluster resource summary through
`infrastructure-monitor-coord`, `app-cluster-api` and `infrastructure-monitor-slave`. `infrastructure-monitor-coord`
also exposes the platform statistics and generic query endpoints for internal usage, which
are routed to the requested cluster through `app-cluster-api`.

## Usage

### `infrastructure-monitor-coord`

```
$ ./bin/infrastructure-monitor-coord run --help
Launch the server API

Usage:
  infrastructure-monitor-coord run [flags]

Flags:
      --appClusterPort int          Port used by app-cluster-api (default 443)
      --appClusterPrefix string     Prefix for application cluster hostnames (default "appcluster")
      --caCert string               Alternative certificate file to use for validation
  -h, --help                        help for run
      --insecure                    Don't validate TLS certificates
      --port int                    Port for Infrastructure Monitor Coordinator gRPC API (default 8423)
      --systemModelAddress string   System Model address (host:port) (default "localhost:8800")
      --useTLS                      Use TLS to connect to application cluster (default true)

Global Flags:
      --consoleLogging   Pretty print logging
      --debug            Set debug level
```

### `infrastructure-monitor-slave`

```
$ ./bin/infrastructure-monitor-slave run --help
Launch the server API

Usage:
  infrastructure-monitor-slave run [flags]

Flags:
  -h, --help                help for run
      --in-cluster          Running inside Kubernetes cluster (--kubeconfig is ignored)
      --kubeconfig string   Kubernetes config file (default "/Users/giel/.kube/config")
      --metricsPort int     Port for HTTP metrics endpoint (default 8424)
      --port int            Port for Infrastructure Monitor Slave gRPC API (default 8422)

Global Flags:
      --consoleLogging   Pretty print logging
      --debug            Set debug level
```

## Integration tests

The following table contains the variables that activate the integration tests

 | Variable  | Example Value | Description |
 | ------------- | ------------- |------------- |
 | `RUN_INTEGRATION_TEST`  | `true` | Run integration tests |
 | `IT_PROMETHEUS_ADDRESS`  | `http://localhost:9090` | Prometheus Address |

To run Prometheus: `docker run --rm -p 9090:9090 prom/prometheus:v2.8.0`

