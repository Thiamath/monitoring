# monitoring

Monitoring the state of application clusters.

## Structure

On each application cluster, `metrics-collector` listens to Kubernetes events and translates
them into Platform-relevant metrics. These are exposed on `:8424/metrics` for Prometheus to scrape.

`metrics-collector` also exposes a gRPC API on port 8422 that provides a cluster resource
summary and platform statistics by querying Prometheus. It also allows for generic queries.

Ultimately, `public-api` provides an interface to retrieve the cluster resource summary through
`monitoring-manager`, `app-cluster-api` and `metrics-collector`. `monitoring-manager`
also exposes the platform statistics and generic query endpoints for internal usage, which
are routed to the requested cluster through `app-cluster-api`.

### Prerequisites

Monitoring requires the following components to be up and running:

* [system-model](https://github.com/nalej/system-model)
* [Prometheus](https://prometheus.io/)


### Build and compile

In order to build and compile this repository use the provided Makefile:

```
make all
```

This operation generates the binaries for this repo, downloads the required dependencies, runs existing tests and generates ready-to-deploy Kubernetes files.

### Run tests

Tests are executed using Ginkgo. To run all the available tests:

```
make test
```

### Update dependencies

Dependencies are managed using Godep. For an automatic dependencies download use:

```
make dep
```

In order to have all dependencies up-to-date run:

```
dep ensure -update -v
```

## Usage


### `monitoring-manager`

```
$ ./bin/monitoring-manager run --help
Launch the server API

Usage:
  monitoring-manager run [flags]

Flags:
      --appClusterPort int          Port used by app-cluster-api (default 443)
      --appClusterPrefix string     Prefix for application cluster hostnames (default "appcluster")
      --caCert string               Alternative certificate file to use for validation
  -h, --help                        Help for run
      --skipServerCertValidation    Don't validate TLS certificates
      --port int                    Port for Infrastructure Monitor Coordinator gRPC API (default 8423)
      --systemModelAddress string   System Model address (host:port) (default "localhost:8800")
      --useTLS                      Use TLS to connect to application cluster (default true)

Global Flags:
      --consoleLogging   Pretty print logging
      --debug            Set debug level
```

### `metrics-collector`

```
$ ./bin/metrics-collector run --help
Launches the server API

Usage:
  metrics-collector run [flags]

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

The following table contains the variables that activate the integration tests:

 | Variable  | Example Value | Description |
 | ------------- | ------------- |------------- |
 | `RUN_INTEGRATION_TEST`  | `true` | Run the integration tests |
 | `IT_PROMETHEUS_ADDRESS`  | `http://localhost:9090` | Prometheus address |

To run Prometheus: `docker run --rm -p 9090:9090 prom/prometheus:v2.8.0`

# Platform Availability Monitoring

The Platform Availability Monitoring runs in Grafana Cloud.

- Main dashboard: https://nalej.grafana.net.
- Grafana Cloud admin: https://grafana.com/orgs/nalej
- Prometheus: https://prometheus-us-central1.grafana.net
- Prometheus data source: https://prometheus-us-central1.grafana.net/api/prom (access: `proxy`)
- Prometheus `remote_write` endpoint: https://prometheus-us-central1.grafana.net/api/prom/push

The main dashboard uses Grafana Cloud as single sign-on, so you need an account that is
authorized to the Nalej organization. Prometheus uses a username and password combo that can be
found in the admin console as API keys.

## Architecture

On the management and application clusters we run Prometheus, deployed with the Prometheus
Operator (`components/prometheus`). We have a number of recording rules that pre-compute
the information needed for our dashboards and alerting (`components/prometheus/appcluster/prometheus.prometheusrule.yaml`).
For clusters that are monitored, we add a `remote_write` section to the configuration that
pushes the result of the recording rules (and nothing else) to the Prometheus in Grafana Cloud.

In Grafana we have a main dashboard that shows the overall health and a list of nodes and components
that are degraded. Each cluster has a dashboard (in the "Clusters" folder) that gives readiness
information per node, as well as component degradation and performance information. This dashboard
includes alerts for bad situations. Alert actions are set up in Grafana itself; we just mark the
actions that we want as "default" so they get triggered whenever any of the cluster dashboards detects
an anomaly.

## Adding a cluster

To add a cluster, the steps are:

1. Create API key for the cluster.
2. Add `remote_write` section to the Prometheus configuration.
3. Add a cluster-specific dashboard to Prometheus.

This process will be automated, but the section below describes how to do this (mostly) manually, for informational purposes.

### Create API key

We need a `MetricsPublisher` API key per cluster, named with the cluster ID. This can be done
through the Grafana Cloud admin console ("API Keys" under the "Security" heading, "Add API key" button
in the top right) or through the (undocumented) API.

To use the API, first manually create an API key with 'admin' role. We can use this key to create more
keys programmatically:

```
$ curl -H "Authorization: Bearer <admin_api_token>" \ 
       -H "Content-Type: application/json" \
       -X POST \ 
       -d '{"name": "<cluster_id>", "role": "MetricsPublisher"}' \
       https://grafana.com/api/orgs/nalej/api-keys
```

This returns the key as the `token` field in a JSON object.

### Add `remote_write` section

For this, we need to create a secret and then a `remote_write` section in the Prometheus
configuration. The secret needs to contain the username (which is unique to our organization and is
`9179`) and the API key created in the previous section.

```
$ kubectl -nnalej create secret generic --from-literal=username=9179 --from-literal=password=<api_token>
```

Now, retrieve the existing Prometheus config:

```
$ kubectl -nnalej get prometheus k8s -oyaml > /tmp/p.yaml
```

Edit the resulting `p.yaml` by adding the following to the `spec` section (and adding the correct cluster and organization ids):

```
  externalLabels:
    cluster_id: <cluster_id>
    organization_id: <organization_id>
  remoteWrite:
  - basicAuth:
      password:
        key: password
        name: grafana-remote-write
      username:
        key: username
        name: grafana-remote-write
    url: https://prometheus-us-central1.grafana.net/api/prom/push
    writeRelabelConfigs:
    - action: keep
      regex: '[a-zA-Z_]*:[a-zA-Z_]+:[a-zA-Z_]*'
      sourceLabels:
      - __name__
```

With these changes, this cluster will show up in the general dashboard.

### Add cluster dashboard

To enable the alerts, we need to add a cluster dashboard. We have a template for this in `scripts/grafana_cluster_alerting_dashboard_template.json`. This template contains some variables that need to be replaced:

- `{{.OrganizationId}}`
- `{{.ClusterId}}`
- `{{.ClusterPublicHostname}}`
- `{{.ClusterName}}` - for this we can just use the same as `ClusterPublicHostname` for now.

Note that we have a few variables that need to be left alone, as they are for the dashboard functionality themselves: `{{node}} {{deployment}} {{statefulset}} {{daemonset}}`.

This dashboard (after replacing the variables) can be added manually through the Grafana API, or by using the following API:

```
$ curl -H "Authorization: Bearer <grafana_api_token>" \
       -H "Content-Type: application/json" \
       -X POST -d @dashboard.json \
       https://nalej.grafana.net/api/dashboards/db
```

The dashboard with the right values is in this example stored as `dashboard.json`. The API token in the command above is a Grafana API token, which is different from the one we used before. This token can be created in Grafana (not the Grafana
Cloud admin console) under "Settings", and the option "API keys".

Note: the template has a `folderId` which puts the dashboard in the Clusters folder. This ID is valid for the
current instance of our Grafana Cloud. To figure out the right ID in future instances, use the
`https://nalej.grafana.net/api/folders` endpoint (with the Grafana API token).


​
## Contributing
​
Please read [contributing.md](contributing.md) for details on our code of conduct, and the process for submitting pull requests to us.
​
​
## Versioning
​
We use [SemVer](http://semver.org/) for versioning. For the available versions, see the [tags on this repository](https://github.com/nalej/monitoring/tags). 
​
## Authors
​
See also the list of [contributors](https://github.com/nalej/monitoring/contributors) who participated in this project.
​
## License
This project is licensed under the Apache 2.0 License - see the [LICENSE-2.0.txt](LICENSE-2.0.txt) file for details.
