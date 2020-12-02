<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# Karavi Metrics for PowerFlex

Karavi Metrics for PowerFlex is part of the Karavi open source suite of Kubernetes storage enablers for Dell EMC products.

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](docs/CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/github/license/dell/karavi-metrics-powerflex)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/dellemc/karavi-metrics-powerflex)](https://hub.docker.com/r/dellemc/karavi-metrics-powerflex)
[![Go version](https://img.shields.io/github/go-mod/go-version/dell/karavi-metrics-powerflex)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/dell/karavi-metrics-powerflex?label=latest&style=flat-square)](https://github.com/dell/karavi-metrics-powerflex/releases)

Karavi Metrics for PowerFlex is an open source distributed solution that provides standardized approaches to gaining observability into Dell EMC PowerFlex. This project provides the following metrics:

- **[I/O Performance Metrics](./docs/IO_PERFORMANCE.md)**: Visibility into the I/O performance of a storage system (IOPs, bandwidth, latency) broken down by export node and volume
- **[Storage Capacity Metrics](./docs/STORAGE_CAPACITY.md)**: Visibility into the total, used, and available capacity for a storage pool/storage class

Karavi Metrics for PowerFlex captures telemetry data of storage usage and performance obtained through the CSI Driver for Dell EMC PowerFlex. The Metrics service then  pushes it to the OpenTelemetry Collector, so it can be processed, and exported in a format consumable by Prometheus.  Prometheus can then be configured to scrape the OpenTelemetry Collector exporter endpoint to provide metrics so they can be visualized in Grafana. Please see [Getting Started Guide](./docs/GETTING_STARTED_GUIDE.md) for information on requirements, deployment, and usage.

## Supported Dell EMC Products

This project currently supports the following Dell EMC storage systems and associated CSI drivers.

| Dell EMC Storage Product | CSI Driver |
| ----------------------- | ---------- |
| PowerFlex v3.0/3.5 | [CSI Driver for PowerFlex v1.1.5, 1.2.0, 1.2.1](https://github.com/dell/csi-vxflexos) |

## Table of Content

- [Code of Conduct](./docs/CODE_OF_CONDUCT.md)
- Guides
  - [Maintainer Guide](./docs/MAINTAINER_GUIDE.md)
  - [Committer Guide](./docs/COMMITTER_GUIDE.md)
  - [Contributing Guide](./docs/CONTRIBUTING.md)
  - [Getting Started Guide](./docs/GETTING_STARTED_GUIDE.md)
- [List of Adopters](./ADOPTERS.md)
- [Maintainers](./docs/MAINTAINERS.md)
- [Support](./docs/SUPPORT.md)
- [Security](./docs/SECURITY.md)
- [About](#about)

## Support

Don’t hesitate to ask! Contact the team and community on [our support page](./docs/SUPPORT.md).
Open an issue if you found a bug on [Github Issues](https://github.com/dell/karavi-metrics-powerflex/issues).

## Versioning

This project is adhering to [Semantic Versioning](https://semver.org/).

## About

Karavi is 100% open source and community-driven. All components are available
under [Apache 2 License](https://www.apache.org/licenses/LICENSE-2.0.html) on
GitHub.
