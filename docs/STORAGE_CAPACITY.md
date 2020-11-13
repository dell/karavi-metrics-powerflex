# Storage Capacity
Provides visibility into the total, used, and available capacity for a storage class and associated underlying storage construct.

To disable these metrics, set the ```storage_class_pool_metrics_enabled``` field to false in helm/values.yaml.

The [Grafana reference dashboards](../../../grafana/dashboards/powerflex) for storage capacity/consumption can be uploaded to your Grafana instance.

## Available Metrics from the OpenTelemetry Collector
The following metrics are available from the OpenTelemetry collector endpoint.  Please see the [GETTING STARTED GUIDE](../GETTING_STARTED_GUIDE.md) for more information on deploying and configuring the OpenTelemetry collector.

### PowerFlex Metrics

| Metric                                       | Description                                                                   | Example                                                                                                                                                                               |
| -------------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| powerflex_storage_pool_total_logical_capacity_gigabytes     | The logical capacity (size) of a storage pool (GB)                            | powerflex_storage_pool_total_logical_capacity_gigabytes{driver="csi-vxflexos.dellemc.com",storage_class="vxflexos",storage_pool="mypool",storage_system_name="2e8ef5244898a20f"} 268.51708984375         |
| powerflex_storage_pool_logical_capacity_available_gigabytes | The capacity available for use (GB)                                           | powerflex_storage_pool_logical_capacity_available_gigabytes{driver="csi-vxflexos.dellemc.com",storage_class="vxflexos-xfs",storage_pool="mypool",storage_system_name="2e8ef5244898a20f"} 253.49462890625 |
| powerflex_storage_pool_logical_capacity_in_use_gigabytes     | The logical capacity of a storage pool in use (GB)                            | powerflex_storage_pool_logical_capacity_in_use_gigabytes{driver="csi-vxflexos.dellemc.com",storage_class="vxflexos-xfs",storage_pool="mypool",storage_system_name="2e8ef5244898a20f"} 15.0224609375       |
| powerflex_storage_pool_logical_provisioned_gigabytes       | The total size of volumes (thick and thin) provisioned in a storage pool (GB) | powerflex_storage_pool_logical_provisioned_gigabytes{driver="csi-vxflexos.dellemc.com",storage_class="vxflexos-xfs",storage_pool="mypool",storage_system_name="2e8ef5244898a20f"} 96                    |