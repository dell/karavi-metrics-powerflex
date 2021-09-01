<!--
Copyright (c) 2021 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Dell EMC Container Storage Modules (CSM) for Observability - Metrics for PowerFlex

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](https://github.com/dell/csm/blob/main/docs/CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/github/license/dell/karavi-metrics-powerflex)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/dellemc/csm-metrics-powerflex)](https://hub.docker.com/r/dellemc/csm-metrics-powerflex)
[![Go version](https://img.shields.io/github/go-mod/go-version/dell/karavi-metrics-powerflex)](go.mod)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/dell/karavi-metrics-powerflex?include_prereleases&label=latest&style=flat-square)](https://github.com/dell/karavi-metrics-powerflex/releases/latest)

Metrics for PowerFlex is part of Dell EMC Container Storage Modules (CSM) for Observability, which provides Kubernetes administrators standardized approaches for storage observability in Kuberenetes environments.

Metrics for PowerFlex is an open source distributed solution that provides insight into storage usage and performance as it relates to the CSI (Container Storage Interface) Driver for Dell EMC PowerFlex.

Metrics for PowerFlex captures telemetry data of storage usage and performance obtained through the CSI Driver for Dell EMC PowerFlex. The Metrics service pushes it to the OpenTelemetry Collector, so it can be processed, and exported in a format consumable by Prometheus. Prometheus can then be configured to scrape the OpenTelemetry Collector exporter endpoint to provide metrics so they can be visualized in Grafana.

For documentation, please visit [Container Storage Modules documentation](https://dell.github.io/csm-docs/).

## Table of Contents

- [Code of Conduct](https://github.com/dell/csm/blob/main/docs/CODE_OF_CONDUCT.md)
- [Maintainer Guide](https://github.com/dell/csm/blob/main/docs/MAINTAINER_GUIDE.md)
- [Committer Guide](https://github.com/dell/csm/blob/main/docs/COMMITTER_GUIDE.md)
- [Contributing Guide](https://github.com/dell/csm/blob/main/docs/CONTRIBUTING.md)
- [Branching Strategy](https://github.com/dell/csm/blob/main/docs/BRANCHING.md)
- [List of Adopters](https://github.com/dell/csm/blob/main/ADOPTERS.md)
- [Maintainers](https://github.com/dell/csm/blob/main/docs/MAINTAINERS.md)
- [Support](https://github.com/dell/csm/blob/main/docs/SUPPORT.md)
- [Security](https://github.com/dell/csm/blob/main/docs/SECURITY.md)
- [About](#about)

## Building Metrics for PowerFlex

If you wish to clone and build the Metrics for PowerFlex service, a Linux host is required with the following installed:

| Component       | Version   | Additional Information                                                                                                                     |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| Docker          | v19+      | [Docker installation](https://docs.docker.com/engine/install/)                                                                                                    |
| Docker Registry |           | Access to a local/corporate [Docker registry](https://docs.docker.com/registry/)                                                           |
| Golang          | v1.14+    | [Golang installation](https://github.com/travis-ci/gimme)                                                                                                         |
| gosec           |           | [gosec](https://github.com/securego/gosec)                                                                                                          |
| gomock          | v.1.4.3   | [Go Mock](https://github.com/golang/mock)                                                                                                             |
| git             | latest    | [Git installation](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)                                                                              |
| gcc             |           | Run ```sudo apt install build-essential```                                                                                                 |
| kubectl         | 1.18-1.20 | Ensure you copy the kubeconfig file from the Kubernetes cluster to the linux host. [kubectl installation](https://kubernetes.io/docs/tasks/tools/install-kubectl/) |
| Helm            | v.3.3.0   | [Helm installation](https://helm.sh/docs/intro/install/)                                                                                                        |

Once all prerequisites are on the Linux host, follow the steps below to clone and build the metrics service:

1. Clone the repository using the following command: `git clone https://github.com/dell/karavi-metrics-powerflex.git`
1. Set the DOCKER_REPO environment variable to point to the local Docker repository, for example: `export DOCKER_REPO=<ip-address>:<port>`
1. In the karavi-metrics-powerflex directory, run the following command to build the Docker image called karavi-metrics-powerflex: `make clean build docker`
1. Tag (with the "latest" tag) and push the image to the local Docker repository by running the following command: `make tag push`

__Note:__ Linux support only. If you are using a local insecure docker registry, ensure you configure the insecure registries on each of the Kubernetes worker nodes to allow access to the local docker repository.

## Testing Metrics for PowerFlex

From the root directory where the repo was cloned, the unit tests can be executed using the following command:

```console
make test
```

This will also provide code coverage statistics for the various Go packages.

## Support

For all your support needs or to follow the latest ongoing discussions and updates, join our Slack group. Click [Here](http://del.ly/Slack_request) to request your invite.

You can also interact with us on [GitHub](https://github.com/dell/csm) by creating a [GitHub Issue](https://github.com/dell/csm/issues).

## Versioning

This project is adhering to [Semantic Versioning](https://semver.org/).

## About

Dell EMC Container Storage Modules (CSM) is 100% open source and community-driven. All components are available
under [Apache 2 License](https://www.apache.org/licenses/LICENSE-2.0.html) on
GitHub.
