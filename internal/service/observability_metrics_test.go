/*
 Copyright (c) 2026 Dell Inc. or its subsidiaries. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package service_test

import (
	"testing"

	"github.com/dell/karavi-metrics-powerflex/internal/service"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gatherPFLXObsMetric gathers metrics from the registry and returns the named MetricFamily.
func gatherPFLXObsMetric(t *testing.T, reg *prometheus.Registry, name string) *dto.MetricFamily {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	for _, mf := range mfs {
		if mf.GetName() == name {
			return mf
		}
	}
	return nil
}

// pflxLabelsMatch checks that all expected key/value pairs are present in the label set.
func pflxLabelsMatch(got []*dto.LabelPair, want map[string]string) bool {
	index := make(map[string]string, len(got))
	for _, lp := range got {
		index[lp.GetName()] = lp.GetValue()
	}
	for k, v := range want {
		if index[k] != v {
			return false
		}
	}
	return true
}

// counterPFLXObs returns the value of the first counter matching the given labels.
func counterPFLXObs(mf *dto.MetricFamily, labels map[string]string) (float64, bool) {
	for _, m := range mf.GetMetric() {
		if pflxLabelsMatch(m.GetLabel(), labels) {
			return m.GetCounter().GetValue(), true
		}
	}
	return 0, false
}

// gaugePFLXObs returns the value of the first gauge matching the given labels.
func gaugePFLXObs(mf *dto.MetricFamily, labels map[string]string) (float64, bool) {
	for _, m := range mf.GetMetric() {
		if pflxLabelsMatch(m.GetLabel(), labels) {
			return m.GetGauge().GetValue(), true
		}
	}
	return 0, false
}

// histogramPFLXObsCount returns the sample count of the first histogram matching the given labels.
func histogramPFLXObsCount(mf *dto.MetricFamily, labels map[string]string) (uint64, bool) {
	for _, m := range mf.GetMetric() {
		if pflxLabelsMatch(m.GetLabel(), labels) {
			return m.GetHistogram().GetSampleCount(), true
		}
	}
	return 0, false
}

// U-OBS-PFLX-01: RecordsAllSelfMetrics registers and records all four self-metrics.
func TestPFLXObsInstrumenter_RecordsAllSelfMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	inst := service.NewPFLXObsInstrumenter(reg)

	inst.RecordCollectionRate("system-abc123", 2.5)
	inst.RecordExportSuccess("system-abc123", "success")
	inst.SetArrayConnectivity("system-abc123", true)
	inst.RecordProcessingLatency("system-abc123", 0.25)

	labels := map[string]string{"module": "metrics-powerflex", "system_id": "system-abc123"}

	mfRate := gatherPFLXObsMetric(t, reg, "dell_csm_obs_collection_rate")
	require.NotNil(t, mfRate, "dell_csm_obs_collection_rate must be registered")
	v, ok := gaugePFLXObs(mfRate, labels)
	assert.True(t, ok, "collection_rate metric must have expected labels")
	assert.Equal(t, 2.5, v, "collection rate must match")

	mfExport := gatherPFLXObsMetric(t, reg, "dell_csm_obs_export_success_total")
	require.NotNil(t, mfExport, "dell_csm_obs_export_success_total must be registered")
	exportLabels := map[string]string{"module": "metrics-powerflex", "system_id": "system-abc123", "status": "success"}
	cv, ok := counterPFLXObs(mfExport, exportLabels)
	assert.True(t, ok, "export_success counter must have expected labels")
	assert.Equal(t, 1.0, cv, "export success counter must be 1 after one call")

	mfConn := gatherPFLXObsMetric(t, reg, "dell_csm_obs_array_connectivity")
	require.NotNil(t, mfConn, "dell_csm_obs_array_connectivity must be registered")
	connVal, ok := gaugePFLXObs(mfConn, labels)
	assert.True(t, ok, "connectivity gauge must have expected labels")
	assert.Equal(t, 1.0, connVal, "connected array must report 1")

	mfLatency := gatherPFLXObsMetric(t, reg, "dell_csm_obs_processing_latency_seconds")
	require.NotNil(t, mfLatency, "dell_csm_obs_processing_latency_seconds must be registered")
	latencyCount, ok := histogramPFLXObsCount(mfLatency, labels)
	assert.True(t, ok, "latency histogram must have expected labels")
	assert.Equal(t, uint64(1), latencyCount, "latency histogram must have one sample")
}

// U-OBS-PFLX-02: SetArrayConnectivity reports 0 for disconnected arrays.
func TestPFLXObsInstrumenter_DisconnectedArray(t *testing.T) {
	reg := prometheus.NewRegistry()
	inst := service.NewPFLXObsInstrumenter(reg)

	inst.SetArrayConnectivity("system-xyz789", false)

	labels := map[string]string{"module": "metrics-powerflex", "system_id": "system-xyz789"}
	mfConn := gatherPFLXObsMetric(t, reg, "dell_csm_obs_array_connectivity")
	require.NotNil(t, mfConn)
	v, ok := gaugePFLXObs(mfConn, labels)
	assert.True(t, ok)
	assert.Equal(t, 0.0, v, "disconnected array must report 0")
}

// U-OBS-PFLX-03: nil receiver does not panic on any method.
func TestPFLXObsInstrumenter_NilReceiverDoesNotPanic(t *testing.T) {
	var inst *service.PFLXObsInstrumenter

	assert.NotPanics(t, func() { inst.RecordCollectionRate("system-abc123", 1.0) })
	assert.NotPanics(t, func() { inst.RecordExportSuccess("system-abc123", "success") })
	assert.NotPanics(t, func() { inst.SetArrayConnectivity("system-abc123", true) })
	assert.NotPanics(t, func() { inst.RecordProcessingLatency("system-abc123", 0.1) })
}
