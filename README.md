<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->
# Karavi PowerFlex Metrics

Karavi PowerFlex Metrics is part of the Karavi open source suite of Kubernetes storage enablers for Dell EMC products.

[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg)](docs/CODE_OF_CONDUCT.md) 
[![License](https://img.shields.io/github/license/dell/karavi-powerflex-metrics)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/dellemc/karavi-powerflex-metrics)](https://hub.docker.com/r/dellemc/karavi-powerflex-metrics)
[![Go version](https://img.shields.io/github/go-mod/go-version/dell/karavi-powerflex-metrics)](go.mod)
[![Latest Release](https://img.shields.io/github/v/release/dell/karavi-powerflex-metrics?label=latest&style=flat-square)](https://github.com/dell/karavi-powerflex-metrics/releases)

Karavi PowerFlex Metrics is an open source distributed solution that provides standardized approaches to gaining observability into Dell EMC products. Karavi PowerFlex Metrics provides the following metrics:
- **[I/O Performance Metrics](./docs/IO_PERFORMANCE.md)**: Visibility into the I/O performance of a storage system (IOPs, bandwidth, latency) broken down by export node and volume 
- **[Storage Capacity Metrics](./docs/STORAGE_CAPACITY.md)**: Visibility into the total, used, and available capacity for a storage pool/storage class

Karavi PowerFlex Metrics captures telemetry data about storage usage and performance and pushes it to the OpenTelemetry Collector, so it can be processed, and exported in an open-source telemetry data format of your choice.  You can then configure an observability tool of your choice, such as Prometheus, to scrape the OpenTelemetry Collector exporter endpoint. Please see [Getting Started Guide](./docs/GETTING_STARTED_GUIDE.md) for information on requirements, deployment, and usage.

## Supported Dell EMC Products

Karavi PowerFlex Metrics currently has support for the following Dell EMC storage systems and associated CSI drivers.

| Dell EMC Storage Product | CSI Driver | Kubernetes | OpenTelemetry |
| ----------------------- | ---------- | ---------- | ------------- |
| PowerFlex v3.0/3.5 | [CSI Driver for PowerFlex v1.1.5, 1.2.0](https://github.com/dell/csi-vxflexos) | 1.17.12, 1.18.10, 1.19.3 | 0.9.0 |

## Table of Content
- [Code of Conduct](./docs/CODE_OF_CONDUCT.md)
- Guides
  - [Maintainer Guide](./docs/MAINTAINER_GUIDE.md)
  - [Committer Guide](./docs/COMMITTER_GUIDE.md)
  - [Contributing Guide](./docs/CONTRIBUTING.md)
  - [Getting Started Guide](./docs/GETTING_STARTED_GUIDE.md)
- [List of Adopters](./ADOPTERS.md)
- [Maintainers](./docs/MAINTAINERS.md)
- [Release Notes](./docs/RELEASE_NOTES.md)
- [Support](./docs/SUPPORT.md)
- [Security](./docs/SECURITY.md)
- [About](#about)

## Support

Don’t hesitate to ask! Contact the team and community on the [mailing lists](https://group).
Open an issue if you found a bug on [Github Issues](https://github.com/dell/karavi-powerflex-metrics/issues).

## About

Karavi is 100% open source and community-driven. All components are available
under [Apache 2 License](https://www.apache.org/licenses/LICENSE-2.0.html) on
GitHub.
