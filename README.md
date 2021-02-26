<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Karavi Metrics for PowerFlex

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](docs/CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/github/license/dell/karavi-metrics-powerflex)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/dellemc/karavi-metrics-powerflex)](https://hub.docker.com/r/dellemc/karavi-metrics-powerflex)
[![Go version](https://img.shields.io/github/go-mod/go-version/dell/karavi-metrics-powerflex)](go.mod)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/dell/karavi-metrics-powerflex?include_prereleases&label=latest&style=flat-square)](https://github.com/dell/karavi-metrics-powerflex/releases/latest)

Karavi Metrics for PowerFlex is part of Karavi Observability storage enabler, which provides Kubernetes administrators standardized approaches for storage observability in Kuberenetes environments.

Karavi Metrics for PowerFlex is an open source distributed solution that provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell EMC PowerFlex. This project provides the following metrics:

- **[Storage System I/O Performance Metrics](./docs/IO_PERFORMANCE.md)**: Visibility into the I/O performance of a storage system (IOPs, bandwidth, latency) broken down by Kubernetes node and volume
- **[Storage Pool Consumption By CSI Driver](./docs/STORAGE_CAPACITY.md)**: Visibility into the total, used, and available capacity for a storage pool/storage class

Karavi Metrics for PowerFlex captures telemetry data of storage usage and performance obtained through the CSI Driver for Dell EMC PowerFlex. The Metrics service then  pushes it to the OpenTelemetry Collector, so it can be processed, and exported in a format consumable by Prometheus. Prometheus can then be configured to scrape the OpenTelemetry Collector exporter endpoint to provide metrics so they can be visualized in Grafana. Please see [Getting Started Guide](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md) for information on requirements, deployment, and usage.

## Table of Contents

- [Code of Conduct](https://github.com/dell/karavi-observability/blob/main/docs/CODE_OF_CONDUCT.md)
- Guides
  - [Maintainer Guide](https://github.com/dell/karavi-observability/blob/main/docs/MAINTAINER_GUIDE.md)
  - [Committer Guide](https://github.com/dell/karavi-observability/blob/main/docs/COMMITTER_GUIDE.md)
  - [Contributing Guide](https://github.com/dell/karavi-observability/blob/main/docs/CONTRIBUTING.md)
  - [Getting Started Guide](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md)
  - [Branching Strategy](./docs/BRANCHING.md)
- [List of Adopters](https://github.com/dell/karavi-observability/blob/main/ADOPTERS.md)
- [Maintainers](./docs/MAINTAINERS.md)
- [Support](https://github.com/dell/karavi-observability/blob/main/docs/SUPPORT.md)
- [Security](./docs/SECURITY.md)
- [About](#about)

## Building the Service

If you wish to clone and build the Karavi Metrics for PowerFlex service, a Linux host is required with the following installed:

| Component       | Version   | Additional Information                                                                                                                     |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Docker          | v19+      | [Docker installation](https://docs.docker.com/engine/install/)                                                                                                    |
| Docker Registry |           | Access to a local/corporate [Docker registry](https://docs.docker.com/registry/)                                                           |
| Golang          | v1.14+    | [Golang installation](https://github.com/travis-ci/gimme)                                                                                                         |
| gosec           |           | [gosec](https://github.com/securego/gosec)                                                                                                          |
| gomock          | v.1.4.3   | [Go Mock](https://github.com/golang/mock)                                                                                                             |
| git             | latest    | [Git installation](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)                                                                              |
| gcc             |           | Run ```sudo apt install build-essential```                                                                                                 |
| kubectl         | 1.17-1.19 | Ensure you copy the kubeconfig file from the Kubernetes cluster to the linux host. [kubectl installation](https://kubernetes.io/docs/tasks/tools/install-kubectl/) |
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

## Support

Don’t hesitate to ask! Contact the team and community on [our support page](https://github.com/dell/karavi-observability/blob/main/docs/SUPPORT.md).
Open an issue if you found a bug on [Github Issues](https://github.com/dell/karavi-observability/issues).

## Versioning

This project is adhering to [Semantic Versioning](https://semver.org/).

## About

Karavi is 100% open source and community-driven. All components are available
under [Apache 2 License](https://www.apache.org/licenses/LICENSE-2.0.html) on
GitHub.

