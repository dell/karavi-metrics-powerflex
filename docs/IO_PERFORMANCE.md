<!--
Copyright (c) 2020 Dell Inc., or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0
-->

# I/O Performance

Storage system I/O performance metrics (IOPS, bandwidth, latency) are available by default and broken down by export node and volume.

To disable these metrics, set the ```sdc_metrics_enabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](../grafana/dashboards/powerflex) for I/O metrics can but uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector

The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [Getting Started Guide](https://github.com/dell/karavi-observability/blob/main/docs/GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerFlex Metrics

| Metric                          | Description                                                             | Example                                                                                                                                                                                                 |
| ------------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| powerflex_export_node_read_bw_megabytes_per_second       | The export node read bandwidth (MB/s) within PowerFlex system                                 | powerflex_export_node_read_bw_megabytes_per_second{ID="cef26c3400000003",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="F5514F1A-C629-4985-8512-A38BBA52882D"} 27.8662109375                                                |
| powerflex_export_node_write_bw_megabytes_per_second      | The export node write bandwidth (MB/s)                                  | powerflex_export_node_write_bw_megabytes_per_second{ID="90c860ec00000001",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="8C911318-9AA9-48B3-A57A-271397B055CF"} 28.248046875                                            |
| powerflex_export_node_read_latency_milliseconds  | The time (in ms) to complete read operations within PowerFlex system by the export node       | powerflex_export_node_read_latency_milliseconds{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 9.648234898015737                                   |
| powerflex_export_node_write_latency_milliseconds | The time (in ms) to complete write operations within PowerFlex system by the export host      | powerflex_export_node_write_latency_milliseconds{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 39.54168373571381                                  |
| powerflex_export_node_read_iops_per_second     | The number of read operations performed by an export node (per second)  | powerflex_export_node_read_iops_per_second{ID="90c860ec00000001",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="8C911318-9AA9-48B3-A57A-271397B055CF"} 1736.6                                                 |
| powerflex_export_node_write_iops_per_second    | The number of write operations performed by an export node (per second) | powerflex_export_node_write_iops_per_second{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 2065                                                  |
| powerflex_volume_read_bw_megabytes_per_second           | The volume read bandwidth (MB/s)                                        | powerflex_volume_read_bw_megabytes_per_second{MappedNodeIDs="\_\_90c860ed00000002\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314200000001",VolumeName="k8s-bf208bb47b"} 21.1630859375           |
| powerflex_volume_write_bw_megabytes_per_second          | The volume write bandwidth (MB/s)                                       | powerflex_volume_write_bw_megabytes_per_second{MappedNodeIDs="\_\_90c860ed00000002__",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 12.484375                |
| powerflex_volume_read_latency_milliseconds      | The time (in ms) to complete read operations to a volume                | powerflex_volume_read_latency_milliseconds{MappedNodeIDs="\_\_90c860ed00000002__",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 7.589428125          |
| powerflex_volume_write_latency_milliseconds     | The time (in ms) to complete write operations to a volume               | powerflex_volume_write_latency_milliseconds{MappedNodeIDs="\_\_90c860ec00000001\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314300000002",VolumeName="k8s-6cb59bdd5c"} 19.65592616580311 |
| powerflex_volume_read_iops_per_second         | The number of read operations performed against a volume (per second)   | powerflex_volume_read_iops_per_second{MappedNodeIDs="\_\_90c860ed00000002\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 753.4                 |
| powerflex_volume_write_iops_per_second        | The number of write operations performed against a volume (per second)  | powerflex_volume_write_iops_per_second{MappedNodeIDs="\_\_90c860ec00000001\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314300000002",VolumeName="k8s-6cb59bdd5c"} 894.4                |
