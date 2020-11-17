<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Getting Started Guide

This document steps through the deployment and configuration of Karavi PowerFlex Metrics.

## Prerequisites

The following are prerequisites for building and deploying Karavi PowerFlex Metrics.

### Kubernetes

The following environment prerequisites are required for deploying karavi-powerflex-metrics.

A Kubernetes cluster with the appropriate version below is required for Karavi PowerFlex Metrics

| Version   | 
| --------- |
| 1.17,1.18,1.19 |

### OpenTelemetry Collector

karavi-powerflex-metrics requires the OpenTelemetry Collector to push metrics to so that the metrics can be consumed by a backend. The helm chart takes care of deploying the Open Telemetry Collector and securing communication between karavi-powerflex-metrics and the Open Telemetry Collector using TLS 1.2 via the user-provided certificate and key files.

Note that although the Open Telmetry Collector can provide metrics for different backends, we currently only support Prometheus.

The OpenTelemetry Collector endpoint to be scraped by Prometheus within the Kubernetes cluster, `otel-collector:443`, is also configured securely (https) via the user-provided certificate and key files. The Prometheus configuration must account for this. See the Prometheus section for an example of a working, minmimal configuration.

### Prometheus

The [Grafana metrics dashboards](../grafana/dashboards/powerflex) require Prometheus to scrape the metrics data from the Open Telemetry Collector.

The Prometheus service should be running on the same Kubernetes cluster as the karavi-powerflex-metrics service and be configured to scrape the OpenTelemetry Collector.

| Supported Version | Image                   | Helm Chart                                                   |
| ----------------- | ----------------------- | ------------------------------------------------------------ |
| 2.22.0           | prom/prometheus:v2.22.0 | https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus |

See below for an example of a working, minmimal configuration. For more information about Prometheus configuration, see https://prometheus.io/docs/prometheus/latest/configuration/configuration/#configuration.

```
scrape_configs:
    - job_name: 'karavi-powerflex-metrics'
      scrape_interval: 5s
      scheme: https
      static_configs:
        - targets: ['otel-collector:443']
      tls_config:
        insecure_skip_verify: true
```

Note that although the Open Telmetry Collector can provide metrics for different backends, we currently only support Prometheus.

#### Grafana

The [Grafana metrics dashboards](../grafana/dashboards/powerflex) require the following Grafana to be deployed in the same Kubernetes cluster as the karavi-powerflex-metrics service. You must also have Prometheus and the OpenTelemetry Collector deployed (see above).

| Supported Version | Helm Chart                                                |
| ----------------- | --------------------------------------------------------- |
| 7.3.0-7.3.2       | https://github.com/grafana/helm-charts/tree/main/charts/grafana |

- Grafana must be configured with the following data sources/plugins:

| Name                   | Additional Information                                                     |
| ---------------------- | -------------------------------------------------------------------------- |
| Prometheus data source | https://grafana.com/docs/grafana/latest/features/datasources/prometheus/   |
| Data Table plugin      | https://grafana.com/grafana/plugins/briangann-datatable-panel/installation |
| Pie Chart plugin       | https://grafana.com/grafana/plugins/grafana-piechart-panel                 |

- Configure the Grafana Prometheus data source

| Setting | Value                     | Additional Information                          |
| ------- | ------------------------- | ----------------------------------------------- |
| Name    | Prometheus                |                                                 |
| Type    | prometheus                |                                                 |
| URL     | http://PROMETHEUS_IP:PORT | The IP/PORT of your running Prometheus instance |
| Access  | Proxy                     |                                                 |

## Building Karavi PowerFlex Metrics (Linux Only)

If you wish to clone and build karavi-powerflex-metrics, a Linux host is required with the following installed:

| Component       | Version   | Additional Information                                                                                                                     |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Docker          | v19+      | https://docs.docker.com/engine/install/                                                                                                    |
| Docker Registry |           | Access to a local/corporate [Docker registry](https://docs.docker.com/registry/)                                                           |
| Golang          | v1.14+    | https://github.com/travis-ci/gimme                                                                                                         |
| gosec           |           | https://github.com/securego/gosec                                                                                                          |
| gomock          | v.1.4.3   | https://github.com/golang/mock                                                                                                             |
| git             | latest    | https://git-scm.com/book/en/v2/Getting-Started-Installing-Git                                                                              |
| gcc             |           | Run ```sudo apt install build-essential```                                                                                                 |
| kubectl         | 1.16-1.17 | Ensure you copy the kubeconfig file from the Kubernetes cluster to the linux host. https://kubernetes.io/docs/tasks/tools/install-kubectl/ |
| Helm            | v.3.3.0   | https://helm.sh/docs/intro/install/                                                                                                        | 

Once all prerequisites are on the Linux host, follow the steps below to clone and build karavi-powerflex-metrics:

1. Clone the karavi-powerflex-metrics repository: `git clone https://github.com/dell/karavi-powerflex-metrics.git`
1. Set the DOCKER_REPO environment variable to point to the local Docker repository, example: `export DOCKER_REPO=<ip-address>:<port>`
1. In the karavi-powerflex-metrics directory, run the following to build the Docker image called karavi-powerflex-metrics: `make clean build docker`
1. To tag (with the "latest" tag) and push the image to the local Docker repository run the following: `make tag push`

__Note:__ If you are using a local insecure docker registry, ensure you configure the insecure registries on each of the Kubernetes worker nodes to allow access to the local docker repository

## Deploying Karavi PowerFlex Metrics
Karavi PowerFlex Metrics is deployed using Helm.  Usage information and available release versions can be found here: https://github.com/dell/helm-charts/tree/main/charts/karavi-powerflex-metrics.

If you built the Karavi PowerFlex Metrics Docker image and pushed it to a local registry, you can deploy it using the same Helm chart above.  You simply need to override the helm chart value pointing to where the Karavi PowerFlex Metrics image lives.  See https://github.com/dell/helm-charts/tree/main/charts/karavi-powerflex-metrics for more details.

## Testing Karavi PowerFlex Metrics

From the karavi-powerflex-metrics root directory where the repo was cloned, the unit tests can be executed as follows:
```console
$ make test
```
This will also provide code coverage statistics for the various Go packages.
