<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Getting Started Guide

This project captures telemetry data about storage usage and performance and pushes it to the [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector), so it can be processed, and exported in a format consumable by Prometheus.  Prometheus can then be configured to scrape the OpenTelemetry Collector exporter endpoint to provide metrics so they can be visualized in Grafana.  

This document steps through the deployment and configuration of this application.

## Kubernetes

First and foremost, the metrics service requires a Kubernetes cluster that aligns with the supported versions listed below.

| Version   |
| --------- |
| 1.17-1.19 |

## Supported Dell EMC Products

This project currently supports the following versions of Dell EMC PowerFlex.

| Version |
| ---------- |
| v3.0, v3.5 |


## CSI Driver for Dell EMC PowerFlex

This project captures telemetry data about Kubernetes storage usage and performance as it relates the CSI (Container Storage Interface) Driver for Dell EMC PowerFlex. The metrics service requires that the CSI Driver for Dell EMC PowerFlex is deployed in the Kubernetes cluster.

| CSI Driver |
| ---------- |
| [CSI Driver for Dell EMC PowerFlex v1.1.5, 1.2.0, 1.2.1](https://github.com/dell/csi-vxflexos) |

## Deploying Karavi Metrics for PowerFlex

This project is deployed using Helm.  Usage information and available release versions can be found here: [Helm chart](https://github.com/dell/helm-charts/tree/main/charts/karavi-metrics-powerflex).

If you built the Docker image and pushed it to a local registry, you can deploy it using the same Helm chart above.  You simply need to override the helm chart value pointing to where the image lives.  See [Karavi Metrics for PowerFlex Helm chart](https://github.com/dell/helm-charts/tree/main/charts/karavi-metrics-powerflex) for more details.

## Required Components

The following third party components are required to be deployed in the same Kubernetes cluster as the karavi-metrics-powerflex service:

* Prometheus
* Grafana

It is the user's responsibility to deploy these components in the same Kubernetes cluster as the metrics service.  These components must be deployed according to the specifications defined below.

The one exception is the OpenTelemetry Collector.  This is deployed and configured as part of this project's deployment.  This is not the user's responsibility.

### OpenTelemetry Collector

The OpenTelemetry Collector is configured to require all communication happen using TLS.  The deployment options listed below will require a signed certificate file and a signed certificate private key file.

The metrics service requires the OpenTelemetry Collector so that metrics can be pushed and later consumed by a backend. The [Helm chart](https://github.com/dell/helm-charts/tree/main/charts/karavi-metrics-powerflex) takes care of deploying the OpenTelemetry Collector and securing communication between the metrics service and the OpenTelemetry Collector using TLS 1.2 via the user-provided certificate and key files.

**Note**: Although the OpenTelemetry Collector can provide metrics for different backends, we currently only support Prometheus.

The OpenTelemetry Collector endpoint is to be scraped by Prometheus, which must be running within the same Kubernetes cluster.

### Prometheus

The [Grafana metrics dashboards](../grafana/dashboards/powerflex) require Prometheus to scrape the metrics data from the Open Telemetry Collector.

The Prometheus service should be running on the same Kubernetes cluster as the metrics service and be configured to scrape the OpenTelemetry Collector.

| Supported Version | Image                   | Helm Chart                                                   |
| ----------------- | ----------------------- | ------------------------------------------------------------ |
| 2.22.0           | prom/prometheus:v2.22.0 | [Prometheus Helm chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus) |

**Note**: It is the user's responsibility to provide persistent storage for Prometheus if they want to preserve historical data.

Here is a sample minimal configuration for Prometheus. For more information about Prometheus configuration, see [Prometheus configuration](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#configuration).

```yaml
scrape_configs:
    - job_name: 'karavi-metrics-powerflex'
      scrape_interval: 5s
      scheme: https
      static_configs:
        - targets: ['otel-collector:443']
      tls_config:
        insecure_skip_verify: true
```

### Grafana

The [Grafana metrics dashboards](../grafana/dashboards/powerflex) require Grafana to be deployed in the same Kubernetes cluster as the metrics service. You must also have Prometheus and the OpenTelemetry Collector deployed (see above). Configure your Grafana instance after successful deployment of the metrics service.
To add the metrics dashboard to Grafana, log in and click the + icon in the side menu. Then click Import. From here you can upload the JSON file or paste the JSON text directly into the text area.

| Supported Version | Helm Chart                                                |
| ----------------- | --------------------------------------------------------- |
| 7.3.0-7.3.2       | [Grafana Helm chart](https://github.com/grafana/helm-charts/tree/main/charts/grafana) |

Grafana must be configured with the following data sources/plugins:

| Name                   | Additional Information                                                     |
| ---------------------- | -------------------------------------------------------------------------- |
| Prometheus data source | [Prometheus data source](https://grafana.com/docs/grafana/latest/features/datasources/prometheus/)   |
| Data Table plugin      | [Data Table plugin](https://grafana.com/grafana/plugins/briangann-datatable-panel/installation) |
| Pie Chart plugin       | [Pie Chart plugin](https://grafana.com/grafana/plugins/grafana-piechart-panel)                 |

Configure the Grafana Prometheus data source

| Setting | Value                     | Additional Information                          |
| ------- | ------------------------- | ----------------------------------------------- |
| Name    | Prometheus                |                                                 |
| Type    | prometheus                |                                                 |
| URL     | http://PROMETHEUS_IP:PORT | The IP/PORT of your running Prometheus instance |
| Access  | Proxy                     |                                                 |

## Building the Service

If you wish to clone and build the metrics service, a Linux host is required with the following installed:

| Component       | Version   | Additional Information                                                                                                                     |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Docker          | v19+      | [Docker installation](https://docs.docker.com/engine/install/)                                                                                                    |
| Docker Registry |           | Access to a local/corporate [Docker registry](https://docs.docker.com/registry/)                                                           |
| Golang          | v1.14+    | [Golang installation](https://github.com/travis-ci/gimme)                                                                                                         |
| gosec           |           | [gosec](https://github.com/securego/gosec)                                                                                                          |
| gomock          | v.1.4.3   | [Go Mock](https://github.com/golang/mock)                                                                                                             |
| git             | latest    | [Git installation](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)                                                                              |
| gcc             |           | Run ```sudo apt install build-essential```                                                                                                 |
| kubectl         | 1.16-1.17 | Ensure you copy the kubeconfig file from the Kubernetes cluster to the linux host. [kubectl installation](https://kubernetes.io/docs/tasks/tools/install-kubectl/) |
| Helm            | v.3.3.0   | [Helm installation](https://helm.sh/docs/intro/install/)                                                                                                        |

Once all prerequisites are on the Linux host, follow the steps below to clone and build the metrics service:

1. Clone the repository: `git clone https://github.com/dell/karavi-metrics-powerflex.git`
1. Set the DOCKER_REPO environment variable to point to the local Docker repository, example: `export DOCKER_REPO=<ip-address>:<port>`
1. In the karavi-metrics-powerflex directory, run the following to build the Docker image called karavi-metrics-powerflex: `make clean build docker`
1. To tag (with the "latest" tag) and push the image to the local Docker repository run the following: `make tag push`

__Note:__ Linux support only. If you are using a local insecure docker registry, ensure you configure the insecure registries on each of the Kubernetes worker nodes to allow access to the local docker repository

## Testing Karavi Metrics for PowerFlex

From the root directory where the repo was cloned, the unit tests can be executed as follows:

```console
make test
```

This will also provide code coverage statistics for the various Go packages.
