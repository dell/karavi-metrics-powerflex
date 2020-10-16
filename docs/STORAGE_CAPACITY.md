# Storage Capacity
Provides visibility into the total, used, and available capacity for a storage class and associated underlying storage construct.

To disable these metrics, set the ```storage_class_pool_metrics_enabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](../../../grafana/dashboards/powerflex) for storage capacity/consumption can be uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector
The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [GETTING STARTED GUIDE](../GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerFlex Metrics

| Metric                                       | Description                                                                   | Example                                                                                                                                                                               |
| -------------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PowerFlexStoragePoolTotalLogicalCapacity     | The logical capacity (size) of a storage pool (GB)                            | PowerFlexStoragePoolTotalLogicalCapacity{Driver="csi-vxflexos.dellemc.com",StorageClass="vxflexos",StoragePool="mypool",StorageSystemName="2e8ef5244898a20f"} 268.51708984375         |
| PowerFlexStoragePoolLogicalCapacityAvailable | The capacity available for use (GB)                                           | PowerFlexStoragePoolLogicalCapacityAvailable{Driver="csi-vxflexos.dellemc.com",StorageClass="vxflexos-xfs",StoragePool="mypool",StorageSystemName="2e8ef5244898a20f"} 253.49462890625 |
| PowerFlexStoragePoolLogicalCapacityInUse     | The logical capacity of a storage pool in use (GB)                            | PowerFlexStoragePoolLogicalCapacityInUse{Driver="csi-vxflexos.dellemc.com",StorageClass="vxflexos-xfs",StoragePool="mypool",StorageSystemName="2e8ef5244898a20f"} 15.0224609375       |
| PowerFlexStoragePoolLogicalProvisioned       | The total size of volumes (thick and thin) provisioned in a storage pool (GB) | PowerFlexStoragePoolLogicalProvisioned{Driver="csi-vxflexos.dellemc.com",StorageClass="vxflexos-xfs",StoragePool="mypool",StorageSystemName="2e8ef5244898a20f"} 96                    |