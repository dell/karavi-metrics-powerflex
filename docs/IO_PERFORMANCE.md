# I/O Performance
Storage system I/O performance metrics (IOPS, bandwidth, latency) are available by default and broken down by export node and volume.

To disable these metrics, set the ```sdc_metrics_enabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](../../../grafana/dashboards/powerflex) for I/O metrics can but uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector
The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [GETTING STARTED GUIDE](../GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerFlex Metrics

| Metric                          | Description                                                             | Example                                                                                                                                                                                                 |
| ------------------------------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PowerFlexExportNodeReadBW       | The export node read bandwidth (MB/s) within PowerFlex system                                 | PowerFlexExportNodeReadBW{ID="cef26c3400000003",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="F5514F1A-C629-4985-8512-A38BBA52882D"} 27.8662109375                                                |
| PowerFlexExportNodeWriteBW      | The export node write bandwidth (MB/s)                                  | PowerFlexExportNodeWriteBW{ID="90c860ec00000001",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="8C911318-9AA9-48B3-A57A-271397B055CF"} 28.248046875                                            |
| PowerFlexExportNodeReadLatency  | The time (in ms) to complete read operations within PowerFlex system by the export node       | PowerFlexExportNodeReadLatency{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 9.648234898015737                                   |
| PowerFlexExportNodeWriteLatency | The time (in ms) to complete write operations within PowerFlex system by the export host      | PowerFlexExportNodeWriteLatency{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 39.54168373571381                                  |
| PowerFlexExportNodeReadIOPS     | The number of read operations performed by an export node (per second)  | PowerFlexExportNodeReadIOPS{ID="90c860ec00000001",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="8C911318-9AA9-48B3-A57A-271397B055CF"} 1736.6                                                 |
| PowerFlexExportNodeWriteIOPS    | The number of write operations performed by an export node (per second) | PowerFlexExportNodeWriteIOPS{ID="90c860ed00000002",IP="1.2.3.4",Name="",PlotWithMean="No",NodeGUID="E147D16C-1FE1-46A4-8E71-F3A8BC59D76B"} 2065                                                  |
| PowerFlexVolumeReadBW           | The volume read bandwidth (MB/s)                                        | PowerFlexVolumeReadBW{MappedSDCIDs="\_\_90c860ed00000002\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314200000001",VolumeName="k8s-bf208bb47b"} 21.1630859375           |
| PowerFlexVolumeWriteBW          | The volume write bandwidth (MB/s)                                       | PowerFlexVolumeWriteBW{MappedSDCIDs="\_\_90c860ed00000002__",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 12.484375                |
| PowerFlexVolumeReadLatency      | The time (in ms) to complete read operations to a volume                | PowerFlexVolumeReadLatency{MappedSDCIDs="\_\_90c860ed00000002__",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 7.589428125          |
| PowerFlexVolumeWriteLatency     | The time (in ms) to complete write operations to a volume               | PowerFlexVolumeWriteLatency{MappedSDCIDs="\_\_90c860ec00000001\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314300000002",VolumeName="k8s-6cb59bdd5c"} 19.65592616580311 |
| PowerFlexVolumeReadIOPS         | The number of read operations performed against a volume (per second)   | PowerFlexVolumeReadIOPS{MappedSDCIDs="\_\_90c860ed00000002\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314100000000",VolumeName="k8s-9afdd0e199"} 753.4                 |
| PowerFlexVolumeWriteIOPS        | The number of write operations performed against a volume (per second)  | PowerFlexVolumeWriteIOPS{MappedSDCIDs="\_\_90c860ec00000001\_\_",MappedNodeIPs="\_\_1.2.3.4\_\_",PlotWithMean="No",VolumeID="069d314300000002",VolumeName="k8s-6cb59bdd5c"} 894.4                |
