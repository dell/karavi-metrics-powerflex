/*
 Copyright (c) 2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

package service

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/sirupsen/logrus"

	sio "github.com/dell/goscaleio"
	types "github.com/dell/goscaleio/types/v1"
	corev1 "k8s.io/api/core/v1"
)

var _ Service = (*PowerFlexService)(nil)

const (
	// DefaultMaxPowerFlexConnections is the number of workers that can query powerflex  at a time
	DefaultMaxPowerFlexConnections = 10

	ExpectedVolumeHandleProperties = 2
)

// Service contains operations that would be used to interact with a PowerFlex system
//
//go:generate mockgen -destination=mocks/service_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service Service
type Service interface {
	GetSDCs(context.Context, PowerFlexClient, SDCFinder) ([]SdcMetricsRetriever, error)
	GetSDCStatistics(context.Context, []corev1.Node, []SdcMetricsRetriever)
	GetVolumes(context.Context, PowerFlexClient, []SdcMetricsRetriever) ([]*VolumeMetaMetrics, error)
	ExportVolumeStatistics(context.Context, []*VolumeMetaMetrics, VolumeFinder)
	GetStorageClasses(ctx context.Context, client PowerFlexClient, storageClassFinder StorageClassFinder) ([]StorageClassMeta, error)
	GetStoragePoolStatistics(ctx context.Context, storageClassMetas []StorageClassMeta)
	ExportTopologyMetrics(context.Context)
}

type SdcMetricsRetriever interface {
	GetStatisticsGetter() StatisticsGetter
	GetGen() string
	GetSdc() *sio.Sdc
	GetClient() PowerFlexClient
}

// StatisticsGetter supports getting statistics
//
//go:generate mockgen -destination=mocks/statistics_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service StatisticsGetter
type StatisticsGetter interface {
	GetStatistics() (*types.SdcStatistics, error)
	GetVolume() ([]*types.Volume, error)
	FindVolumes() ([]*sio.Volume, error)
	GetVolumeMetrics() ([]*types.SdcVolumeMetrics, error)
}

// VolumeStatisticsGetter supports getting statistics
//
//go:generate mockgen -destination=mocks/volume_statistics_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service VolumeStatisticsGetter
type VolumeStatisticsGetter interface {
	GetVolumeStatistics() (*types.VolumeStatistics, error)
}

// PowerFlexClient contains operations for a powerflex client
//
//go:generate mockgen -destination=mocks/powerflex_client_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service PowerFlexClient
type PowerFlexClient interface {
	Authenticate(*sio.ConfigConnect) (sio.Cluster, error)
	GetInstance(string) ([]*types.System, error)
	FindSystem(string, string, string) (*sio.System, error)
	GetStoragePool(href string) ([]*types.StoragePool, error)
	GetMetrics(string, []string) (*types.MetricsResponse, error)
}

// PowerFlexSystem contains operations for a powerflex system
//
//go:generate mockgen -destination=mocks/powerflex_system_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service PowerFlexSystem
type PowerFlexSystem interface {
	FindSdc(string, string) (*sio.Sdc, error)
}

// PowerFlexService represents the service for getting SDC metrics data for a PowerFlex system
type PowerFlexService struct {
	MetricsWrapper          MetricsRecorder
	MaxPowerFlexConnections int
	Logger                  *logrus.Logger
	VolumeFinder            VolumeFinder
}

// SDCFinder is used to find SDC GUIDs
//
//go:generate mockgen -destination=mocks/sdc_finder_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service SDCFinder
type SDCFinder interface {
	GetSDCGuids() ([]string, error)
}

// StorageClassFinder is used to find storage classes in kubernetes
//
//go:generate mockgen -destination=mocks/storage_class_finder_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service StorageClassFinder
type StorageClassFinder interface {
	GetStorageClasses() ([]k8s.StorageClass, error)
	GetStoragePools(storageClass k8s.StorageClass) []string
}

// VolumeFinder is used to find volume information in kubernetes
//
//go:generate mockgen -destination=mocks/volume_finder_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service VolumeFinder
type VolumeFinder interface {
	GetPersistentVolumes() ([]k8s.VolumeInfo, error)
}

// NodeFinder is a node finder that will query the Kubernetes API for a slice of cluster nodes
//
//go:generate mockgen -destination=mocks/node_finder_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service NodeFinder
type NodeFinder interface {
	GetNodes() ([]corev1.Node, error)
}

type StoragePoolMetricsHandler struct {
	StoragePoolStatisticsGetter StoragePoolStatisticsGetter
	GenType                     string
	Client                      PowerFlexClient
}

func (s StoragePoolMetricsHandler) GetClient() PowerFlexClient {
	return s.Client
}

func (s StoragePoolMetricsHandler) GetStatisticsGetter() StoragePoolStatisticsGetter {
	return s.StoragePoolStatisticsGetter
}

func (s StoragePoolMetricsHandler) GetGen() string {
	return s.GenType
}

var _ StoragePoolMetricsRetriever = (*StoragePoolMetricsHandler)(nil)

var _ SdcMetricsRetriever = (*SdcMetricsHandler)(nil)

// SystemFinder is a function that will be used for finding a PowerFlexSystem by id, name, and href
var SystemFinder = func(client PowerFlexClient, id string, name string, href string) (PowerFlexSystem, error) {
	return client.FindSystem(id, name, href)
}

type StoragePoolMetricsRetriever interface {
	GetStatisticsGetter() StoragePoolStatisticsGetter
	GetClient() PowerFlexClient
	GetGen() string
}

// StoragePoolStatisticsGetter supports getting storage pool statistics
//
//go:generate mockgen -destination=mocks/storage_pool_statistics_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service StoragePoolStatisticsGetter
type StoragePoolStatisticsGetter interface {
	GetStatistics() (*types.Statistics, error)
}

// LeaderElector will elect a leader
//
//go:generate mockgen -destination=mocks/leader_elector_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service LeaderElector
type LeaderElector interface {
	InitLeaderElection(string, string) error
	IsLeader() bool
}

// TopologyMetricsRecord used for holding output of the Topology metric query results
type TopologyMetricsRecord struct {
	topologyMeta *TopologyMeta
	pvAvailable  int64
}

// SdcMetricsHandler is used to get SDC metrics
type SdcMetricsHandler struct {
	Sdc              *sio.Sdc
	StatisticsGetter StatisticsGetter
	GenType          string
	Client           PowerFlexClient
}

// storagePoolMetricsRecord used for holding output of the Storage pool stat query results
type storagePoolMetricsRecord struct {
	ID               string
	CTX              context.Context
	storageClassMeta *StorageClassMeta
	TotalLogicalCapacity, LogicalCapacityAvailable,
	LogicalCapacityInUse, LogicalProvisioned float64
}

// IDedPoolStatisticGetter offers PoolStatisticGetter with its corresponding pool ID
type IDedPoolStatisticGetter struct {
	ID     string
	Getter StoragePoolMetricsRetriever
}

// SDCMetricsRecord used for holding output of the SDC stat query results
type SDCMetricsRecord struct {
	sdcMeta *SDCMeta
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float64
}

// VolumeMetricsRecord used for holding output of the Volume stat query results
type VolumeMetricsRecord struct {
	volumeMeta *VolumeMeta
	readBW, writeBW,
	readIOPS, writeIOPS,
	readLatency, writeLatency float64
}

// GetGenType queries the PowerFlex system for its gen type
func GetGenType(system *sio.System) (string, error) {
	pds, err := system.GetProtectionDomain("")
	if err != nil {
		return "", err
	}
	if len(pds) > 0 {
		return pds[0].GenType, nil
	}
	return "", nil
}

// getMetric retrieves a specific metric value from a slice of metrics
func getMetric(metrics []types.Metric, name string) float64 {
	for _, m := range metrics {
		if m.Name == name {
			return m.Values[0]
		}
	}
	return 0
}

// GetSDCs returns a slice of SDCs
func (s *PowerFlexService) GetSDCs(_ context.Context, client PowerFlexClient, sdcFinder SDCFinder) ([]SdcMetricsRetriever, error) {
	var sdcs []SdcMetricsRetriever
	sdcGUIDs, err := sdcFinder.GetSDCGuids()
	if err != nil {
		return nil, err
	}
	s.Logger.WithField("sdc_guids", sdcGUIDs).Debug("get sdc guids")
	if len(sdcGUIDs) == 0 {
		return sdcs, nil
	}
	systems, err := client.GetInstance("")
	if err != nil {
		return nil, err
	}
	for _, system := range systems {
		s.Logger.WithFields(logrus.Fields{"system_id": system.ID, "system_name": system.Name}).Debug("looking up system")
		sys, err := SystemFinder(client, system.ID, system.Name, "")
		if err != nil {
			return nil, err
		}
		realSystem, err := client.FindSystem(system.ID, system.Name, "")
		if err != nil {
			return nil, err
		}
		genType, err := GetGenType(realSystem)
		if err != nil {
			return nil, err
		}
		for _, sdcGUID := range sdcGUIDs {
			sdc, err := sys.FindSdc("SdcGUID", sdcGUID)
			if err != nil {
				s.Logger.WithField("sdc_guid", sdcGUID).Warn("unable to find SDC with GUID")
			} else {
				s.Logger.WithFields(logrus.Fields{"sdc_guid": sdcGUID}).Debug("found sdc")
				sdcs = append(sdcs, SdcMetricsHandler{
					Sdc:              sdc,
					Client:           client,
					StatisticsGetter: sdc,
					GenType:          genType,
				})
			}
		}
	}
	return sdcs, nil
}

func (s SdcMetricsHandler) GetClient() PowerFlexClient {
	return s.Client
}

func (s SdcMetricsHandler) GetSdc() *sio.Sdc {
	return s.Sdc
}

func (s SdcMetricsHandler) GetStatisticsGetter() StatisticsGetter {
	return s.StatisticsGetter
}

func (s SdcMetricsHandler) GetGen() string {
	return s.GenType
}

// GetSDCMeta returns SDC meta information from a goscaleio SDC.
// This function is exported for direct testing.
func GetSDCMeta(sdc interface{}, nodes []corev1.Node) (*SDCMeta, error) {
	if sdc == nil {
		return nil, fmt.Errorf("nil sdc")
	}

	switch v := sdc.(type) {
	case *sio.Sdc:
		var name string
	loop:
		for _, node := range nodes {
			for _, addr := range node.Status.Addresses {
				if addr.Address == v.Sdc.SdcIP {
					name = node.GetName()
					break loop
				}
			}
		}

		return &SDCMeta{
			Name:    name,
			ID:      v.Sdc.ID,
			IP:      v.Sdc.SdcIP,
			SdcGUID: v.Sdc.SdcGUID,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported sdc type %T", sdc)
	}
}

// GetSDCStatistics records I/O statistics for the given list of SDCs
func (s *PowerFlexService) GetSDCStatistics(ctx context.Context, nodes []corev1.Node, sdcs []SdcMetricsRetriever) {
	start := time.Now()
	defer s.timeSince(start, "GetSDCStatistics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting SDCStatistics")
		return
	}

	if s.MaxPowerFlexConnections == 0 {
		s.Logger.Debug("using DefaultMaxPowerFlexConnections")
		s.MaxPowerFlexConnections = DefaultMaxPowerFlexConnections
	}

	for range s.pushSDCMetrics(ctx, s.gatherSDCMetrics(ctx, nodes, s.sdcServer(sdcs))) {
		// consume the channel until it is empty and closed
	} // revive:disable-line:empty-block
}

// sdcServer will create a channel and push all SDCs into it
func (s *PowerFlexService) sdcServer(sdcs []SdcMetricsRetriever) <-chan SdcMetricsRetriever {
	sdcChan := make(chan SdcMetricsRetriever, len(sdcs))
	go func() {
		for _, sdc := range sdcs {
			sdcChan <- sdc
		}
		close(sdcChan)
	}()
	return sdcChan
}

// gatherSDCMetrics will collect, in parallel, stats against each SDC referenced by 'statGetters'
func (s *PowerFlexService) gatherSDCMetrics(_ context.Context, nodes []corev1.Node, sdcs <-chan SdcMetricsRetriever) <-chan *SDCMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherMetrics")

	ch := make(chan *SDCMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerFlexConnections)

	go func() {
		for sdc := range sdcs {
			wg.Add(1)
			sem <- struct{}{}
			go func(sdc SdcMetricsRetriever) {
				defer func() {
					if r := recover(); r != nil {
						s.Logger.Errorf("Error: %v\n%s", r, debug.Stack())
					}
					<-sem
					wg.Done()
				}()

				sdcMeta, err := GetSDCMeta(sdc.GetSdc(), nodes)
				if err != nil {
					s.Logger.WithError(err).Warn("GetSDCMeta failed")
					return
				}
				if sdcMeta == nil {
					s.Logger.Warn("GetSDCMeta returned nil meta")
					return
				}

				if sdc.GetGen() == types.GenTypeEC {
					stats, err := sdc.GetClient().GetMetrics("sdc", []string{sdc.GetSdc().Sdc.ID})
					if err != nil {
						s.Logger.WithError(err).WithField("sdc", sdcMeta.ID).Error("getting statistics for sdc")
						return
					}
					s.Logger.WithField("sdc_ids_for_metrics", sdc.GetSdc().Sdc.ID).Debug("calling GetMetrics(sdc)")

					if len(stats.Resources) == 0 {
						s.Logger.Warn("No resources found in metrics response for SDC")
						return
					}

					readBW := getMetric(stats.Resources[0].Metrics, "host_read_bandwidth")
					writeBW := getMetric(stats.Resources[0].Metrics, "host_write_bandwidth")
					readIOPS := getMetric(stats.Resources[0].Metrics, "host_read_iops")
					writeIOPS := getMetric(stats.Resources[0].Metrics, "host_write_iops")
					readLatency := getMetric(stats.Resources[0].Metrics, "avg_host_read_latency")
					writeLatency := getMetric(stats.Resources[0].Metrics, "avg_host_write_latency")

					s.Logger.WithFields(logrus.Fields{
						"sdc_meta":        sdcMeta,
						"read_bandwidth":  readBW,
						"write_bandwidth": writeBW,
						"read_iops":       readIOPS,
						"write_iops":      writeIOPS,
						"read_latency":    readLatency,
						"write_latency":   writeLatency,
					}).Debug("sdc metrics")

					ch <- &SDCMetricsRecord{
						sdcMeta: sdcMeta,
						readBW:  readBW, writeBW: writeBW,
						readIOPS: readIOPS, writeIOPS: writeIOPS,
						readLatency: readLatency, writeLatency: writeLatency,
					}
				} else {
					stats, err := sdc.GetStatisticsGetter().GetStatistics()
					if err != nil {
						s.Logger.WithError(err).WithField("sdc", sdcMeta.ID).Error("getting statistics for sdc")
						return
					}

					readBW, writeBW := GetSDCBandwidth(stats)
					readIOPS, writeIOPS := GetSDCIOPS(stats)
					readLatency, writeLatency := GetSDCLatency(stats)
					s.Logger.WithFields(logrus.Fields{
						"sdc_meta":        sdcMeta,
						"read_bandwidth":  readBW,
						"write_bandwidth": writeBW,
						"read_iops":       readIOPS,
						"write_iops":      writeIOPS,
						"read_latency":    readLatency,
						"write_latency":   writeLatency,
					}).Debug("sdc metrics")

					ch <- &SDCMetricsRecord{
						sdcMeta: sdcMeta,
						readBW:  readBW, writeBW: writeBW,
						readIOPS: readIOPS, writeIOPS: writeIOPS,
						readLatency: readLatency, writeLatency: writeLatency,
					}
				}
			}(sdc)
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushSDCMetrics will, in parallel, record stats in 'metrics' using the s.MetricsWrapper
func (s *PowerFlexService) pushSDCMetrics(ctx context.Context, sdcMetrics <-chan *SDCMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushMetrics")

	var wg sync.WaitGroup
	ch := make(chan string)

	go func() {
		for record := range sdcMetrics {
			wg.Add(1)
			go func(mr *SDCMetricsRecord) {
				defer wg.Done()
				if mr == nil {
					s.Logger.WithField("sdc", mr.sdcMeta.ID).Warn("empty statistics for sdc")
					return
				}

				err := s.MetricsWrapper.Record(
					ctx, mr.sdcMeta,
					mr.readBW, mr.writeBW,
					mr.readIOPS, mr.writeIOPS,
					mr.readLatency, mr.writeLatency,
				)

				if err != nil {
					s.Logger.WithError(err).WithField("sdc", mr.sdcMeta.ID).Error("recording statistics for sdc")
				} else {
					ch <- mr.sdcMeta.ID
				}
			}(record)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// getVolumeMetaMetrics returns Volume meta information from a goscaleio Volume.
func getVolumeMetaMetrics(volume interface{}) *VolumeMetaMetrics {
	switch v := volume.(type) {
	case *sio.Volume:
		var sdcsInfo []MappedSDC
		for _, d := range v.Volume.MappedSdcInfo {
			sdcsInfo = append(sdcsInfo,
				MappedSDC{SdcID: d.SdcID, SdcIP: d.SdcIP},
			)
		}
		return &VolumeMetaMetrics{
			Name:       v.Volume.Name,
			ID:         v.Volume.ID,
			MappedSDCs: sdcsInfo,
		}
	default:
		return &VolumeMetaMetrics{
			Name:       "",
			ID:         "",
			MappedSDCs: []MappedSDC{},
		}
	}
}

// GetVolumes returns all unique, mapped volumes in sdcs along with their metadata and metrics
func (s *PowerFlexService) GetVolumes(_ context.Context, client PowerFlexClient, sdcs []SdcMetricsRetriever) ([]*VolumeMetaMetrics, error) {
	var uniqueVolumes []*VolumeMetaMetrics
	visited := make(map[string]bool)

	for _, sdc := range sdcs {
		vols, err := sdc.GetSdc().FindVolumes()
		if err != nil {
			return nil, err
		}

		genType := ""
		if len(vols) > 0 {
			genType = vols[0].Volume.GenType
		}

		if genType == types.GenTypeEC {
			volumeIDs := make([]string, 0, len(vols))
			for _, v := range vols {
				volumeIDs = append(volumeIDs, v.Volume.ID)
			}

			cleanIDs := make([]string, 0, len(volumeIDs))
			for _, id := range volumeIDs {
				if id != "" {
					cleanIDs = append(cleanIDs, id)
				}
			}

			if len(cleanIDs) == 0 {
				s.Logger.Warn("no valid volume IDs found for EC metrics; skipping GetMetrics(volume)")
				return uniqueVolumes, nil
			}

			s.Logger.WithField("volume_ids_for_metrics", cleanIDs).Debug("calling GetMetrics(volume)")
			metrics, err := client.GetMetrics("volume", cleanIDs)
			if err != nil {
				return nil, err
			}

			if len(metrics.Resources) == 0 {
				s.Logger.Warn("No resources found in metrics response for volume")
				return nil, fmt.Errorf("no volume metrics found for volume IDs: %v", cleanIDs)
			}
			volMetrics := make(map[string]types.Resource, len(metrics.Resources))
			for _, m := range metrics.Resources {
				volMetrics[m.ID] = m
			}

			for _, v := range vols {
				volumeMeta := getVolumeMetaMetrics(v)
				if !visited[volumeMeta.ID] {
					s.Logger.WithFields(logrus.Fields{
						"volume_id":   volumeMeta.ID,
						"volume_name": volumeMeta.Name,
						"mapped_sdcs": volumeMeta.MappedSDCs,
					}).Debug("Processing volume")

					if metrics, ok := volMetrics[volumeMeta.ID]; ok {
						s.Logger.WithField("metrics_found", true).Debug("Volume metrics available")

						volumeMeta.HostReadBandwith = getMetric(metrics.Metrics, "host_read_bandwidth")
						volumeMeta.HostWriteBandwith = getMetric(metrics.Metrics, "host_write_bandwidth")
						volumeMeta.HostReadIOPS = getMetric(metrics.Metrics, "host_read_iops")
						volumeMeta.HostWriteIOPS = getMetric(metrics.Metrics, "host_write_iops")
						volumeMeta.AvgHostReadLatency = getMetric(metrics.Metrics, "avg_host_read_latency")
						volumeMeta.AvgHostWriteLatency = getMetric(metrics.Metrics, "avg_host_write_latency")

						// Normalize units for readability
						// Bandwidth: bytes/sec → MB/sec
						volumeMeta.HostReadBandwith = volumeMeta.HostReadBandwith / (1024 * 1024)
						volumeMeta.HostWriteBandwith = volumeMeta.HostWriteBandwith / (1024 * 1024)
						// Latency: microseconds → milliseconds
						volumeMeta.AvgHostReadLatency = volumeMeta.AvgHostReadLatency / 1000
						volumeMeta.AvgHostWriteLatency = volumeMeta.AvgHostWriteLatency / 1000

						// set the GenType
						volumeMeta.GenType = genType

						s.Logger.WithFields(logrus.Fields{
							"read_bw":    volumeMeta.HostReadBandwith,
							"write_bw":   volumeMeta.HostWriteBandwith,
							"read_iops":  volumeMeta.HostReadIOPS,
							"write_iops": volumeMeta.HostWriteIOPS,
							"read_lat":   volumeMeta.AvgHostReadLatency,
							"write_lat":  volumeMeta.AvgHostWriteLatency,
						}).Debug("Volume metrics populated")
					} else {
						s.Logger.WithField("metrics_found", false).Warn("No metrics found for volume")
					}
					uniqueVolumes = append(uniqueVolumes, volumeMeta)
					visited[volumeMeta.ID] = true
				}
			}
		} else {
			metrics, err := sdc.GetStatisticsGetter().GetVolumeMetrics()
			if err != nil {
				return nil, err
			}
			volMetrics := make(map[string]*types.SdcVolumeMetrics, len(metrics))
			for _, m := range metrics {
				volMetrics[m.VolumeID] = m
			}

			for _, v := range vols {
				volumeMeta := getVolumeMetaMetrics(v)
				if !visited[volumeMeta.ID] {
					s.Logger.WithField("volume_id", volumeMeta.ID).Debug("found volume")
					if _, ok := volMetrics[volumeMeta.ID]; ok {
						volumeMeta.ReadBwc = volMetrics[volumeMeta.ID].ReadBwc
						volumeMeta.WriteBwc = volMetrics[volumeMeta.ID].WriteBwc
						volumeMeta.ReadLatencyBwc = volMetrics[volumeMeta.ID].ReadLatencyBwc
						volumeMeta.WriteLatencyBwc = volMetrics[volumeMeta.ID].WriteLatencyBwc
						volumeMeta.TrimBwc = volMetrics[volumeMeta.ID].TrimBwc
						volumeMeta.TrimLatencyBwc = volMetrics[volumeMeta.ID].TrimLatencyBwc
					}
					uniqueVolumes = append(uniqueVolumes, volumeMeta)
					visited[volumeMeta.ID] = true
				}
			}
		}
	}
	return uniqueVolumes, nil
}

// ExportVolumeStatistics records I/O statistics for the given list of Volumes
func (s *PowerFlexService) ExportVolumeStatistics(ctx context.Context, volumes []*VolumeMetaMetrics, volumeFinder VolumeFinder) {
	start := time.Now()
	defer s.timeSince(start, "ExportVolumeStatistics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportVolumeStatistics")
		return
	}

	if s.MaxPowerFlexConnections == 0 {
		s.Logger.Debug("Using DefaultMaxPowerFlexConnections")
		s.MaxPowerFlexConnections = DefaultMaxPowerFlexConnections
	}

	for range s.pushVolumeMetrics(ctx, s.gatherVolumeMetrics(ctx, volumeFinder, s.volumeServer(volumes))) {
		// consume the channel until it is empty and closed
	} // revive:disable-line:empty-block
}

// volumeServer will return a channel of volumes that can provide statistics about each volume
func (s *PowerFlexService) volumeServer(volumes []*VolumeMetaMetrics) <-chan *VolumeMetaMetrics {
	volumeChannel := make(chan *VolumeMetaMetrics, len(volumes))
	go func() {
		for _, volume := range volumes {
			volumeChannel <- volume
		}
		close(volumeChannel)
	}()
	return volumeChannel
}

// volumeServer will return a channel of volumes that can provide  about each volume
func (s *PowerFlexService) getVolumeInfo(_ context.Context, volumes []k8s.VolumeInfo) <-chan k8s.VolumeInfo {
	volumeChannel := make(chan k8s.VolumeInfo, len(volumes))
	go func() {
		for _, volume := range volumes {
			volumeChannel <- volume
		}
		close(volumeChannel)
	}()
	return volumeChannel
}

// gatherVolumeMetrics will return a channel of volume metrics based on the input of volumes
func (s *PowerFlexService) gatherVolumeMetrics(_ context.Context, volumeFinder VolumeFinder, volumes <-chan *VolumeMetaMetrics) <-chan *VolumeMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherVolumeMetrics")

	ch := make(chan *VolumeMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerFlexConnections)

	go func() {
		persistentVolumes := make(map[string]k8s.VolumeInfo)
		pvs, err := volumeFinder.GetPersistentVolumes()
		if err != nil {
			s.Logger.WithError(err).Error("getting persistent volumes")
			return
		}
		for _, v := range pvs {
			persistentVolumes[v.StorageSystemVolumeName] = v
		}

		exported := false
		for volume := range volumes {
			exported = true
			wg.Add(1)
			sem <- struct{}{}
			go func(volume *VolumeMetaMetrics) {
				defer func() {
					wg.Done()
					<-sem
				}()

				if pv, ok := persistentVolumes[volume.Name]; ok {
					volume.PersistentVolumeName = pv.PersistentVolume
					volume.StorageSystemID = pv.StorageSystemID
					volume.Namespace = pv.Namespace
					volume.PersistentVolumeClaimName = pv.VolumeClaimName
				} else {
					s.Logger.WithField("volume_id", volume.ID).Error("could not find a Persistent Volume that maps to storage system volume ID")
				}

				readBW, writeBW := GetVolumeBandwidth(volume)
				readIOPS, writeIOPS := GetVolumeIOPS(volume)
				readLatency, writeLatency := GetVolumeLatency(volume)

				volumeMeta := &VolumeMeta{
					ID:                        volume.ID,
					Name:                      volume.Name,
					PersistentVolumeName:      volume.PersistentVolumeName,
					PersistentVolumeClaimName: volume.PersistentVolumeClaimName,
					Namespace:                 volume.Namespace,
					StorageSystemID:           volume.StorageSystemID,
					MappedSDCs:                volume.MappedSDCs,
				}

				s.Logger.WithFields(logrus.Fields{
					"volume_meta":     volumeMeta,
					"read_bandwidth":  readBW,
					"write_bandwidth": writeBW,
					"read_iops":       readIOPS,
					"write_iops":      writeIOPS,
					"read_latency":    readLatency,
					"write_latency":   writeLatency,
				}).Debug("volume metrics")

				ch <- &VolumeMetricsRecord{
					volumeMeta: volumeMeta,
					readBW:     readBW, writeBW: writeBW,
					readIOPS: readIOPS, writeIOPS: writeIOPS,
					readLatency: readLatency, writeLatency: writeLatency,
				}
			}(volume)
		}

		if !exported {
			// If no volumes metrics were exported, we need to export an "empty" metric to update the OT Collector
			// so that stale entries are removed
			ch <- &VolumeMetricsRecord{
				volumeMeta: &VolumeMeta{},
				readBW:     0, writeBW: 0,
				readIOPS: 0, writeIOPS: 0,
				readLatency: 0, writeLatency: 0,
			}
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushVolumeMetrics will push the provided channel of volume metrics to a data collector
func (s *PowerFlexService) pushVolumeMetrics(ctx context.Context, volumeMetrics <-chan *VolumeMetricsRecord) <-chan string {
	start := time.Now()
	defer s.timeSince(start, "pushVolumeMetrics")
	var wg sync.WaitGroup

	ch := make(chan string)
	go func() {
		for metrics := range volumeMetrics {
			wg.Add(1)
			go func(metrics *VolumeMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.Record(ctx,
					metrics.volumeMeta,
					metrics.readBW, metrics.writeBW,
					metrics.readIOPS, metrics.writeIOPS,
					metrics.readLatency, metrics.writeLatency,
				)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", metrics.volumeMeta.ID).Error("recording statistics for volume")
				} else {
					ch <- metrics.volumeMeta.ID
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// GetStorageClasses returns a list of StorageClassMeta
func (s *PowerFlexService) GetStorageClasses(_ context.Context, client PowerFlexClient, storageClassFinder StorageClassFinder) ([]StorageClassMeta, error) {
	var c *sio.Client
	switch underlyingClient := client.(type) {
	case *sio.Client:
		c = underlyingClient
	default:
		// need a *sio.Client to create a sio storage pool instance on line 287
		// client is mock client during tests so we need to set an *sio.Client
		// should never get here in production
		c = &sio.Client{}
	}

	storageClassMetas := []StorageClassMeta{}
	storageClassInfos := []StorageClassInfo{}

	storageClasses, err := storageClassFinder.GetStorageClasses()
	if err != nil {
		return nil, err
	}

	systems, err := client.GetInstance("")
	if err != nil {
		return nil, err
	}

	if len(systems) == 0 {
		return nil, fmt.Errorf("no systems found")
	}

	s.Logger.WithField("systems", len(systems)).Debug("systems log")
	for _, system := range systems {
		s.Logger.WithField("system", system.ID).Debug("systems log")
		for _, class := range storageClasses {
			systemid := class.SystemID
			s.Logger.WithField("class.Parameters", systemid).Debug("systems log")
			if system.ID == systemid {
				storageClassInfo := StorageClassInfo{
					ID:              string(class.UID),
					Name:            class.Name,
					Driver:          class.Provisioner,
					StorageSystemID: systemid,
					StoragePools:    storageClassFinder.GetStoragePools(class),
				}
				s.Logger.WithField("storage_class_info", storageClassInfo).Debug("found storage class")
				storageClassInfos = append(storageClassInfos, storageClassInfo)
			}
		}
	}

	systemStoragePools, err := client.GetStoragePool("")
	if err != nil {
		return nil, err
	}

	for _, class := range storageClassInfos {
		storageClassMeta := StorageClassMeta{
			ID:              class.ID,
			Name:            class.Name,
			Driver:          class.Driver,
			StorageSystemID: class.StorageSystemID,
			StoragePools:    make(map[string]StoragePoolMetricsRetriever),
		}

		for _, systemPool := range systemStoragePools {
			if contains(class.StoragePools, systemPool.Name) {
				storageClassMeta.StoragePools[systemPool.ID] = StoragePoolMetricsHandler{
					GenType:                     systemPool.GenType,
					StoragePoolStatisticsGetter: sio.NewStoragePoolEx(c, systemPool),
					Client:                      client,
				}
			}
		}

		storageClassMetas = append(storageClassMetas, storageClassMeta)
	}

	return storageClassMetas, nil
}

// GetStoragePoolStatistics records the capacity metrics for a slice of StorageClassMeta
func (s *PowerFlexService) GetStoragePoolStatistics(ctx context.Context, storageClassMetas []StorageClassMeta) {
	start := time.Now()
	defer s.timeSince(start, "GetStoragePoolStatistics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting Storage Pool statistics")
		return
	}
	if s.MaxPowerFlexConnections == 0 {
		s.Logger.Debug("using DefaultMaxPowerFlexConnections")
		s.MaxPowerFlexConnections = DefaultMaxPowerFlexConnections
	}

	for i, storageClassMeta := range storageClassMetas {
		for range s.pushPoolStatistics(ctx, s.gatherPoolStatistics(ctx, &storageClassMetas[i], s.storagePoolServer(storageClassMeta.StoragePools))) {
			// consume the channel until empty and closed
		} // revive:disable-line:empty-block
	}
}

// storagePoolServer will create a channel and push all StoragePools into it
func (s *PowerFlexService) storagePoolServer(pools map[string]StoragePoolMetricsRetriever) <-chan IDedPoolStatisticGetter {
	poolChannel := make(chan IDedPoolStatisticGetter)
	go func() {
		for id, pool := range pools {
			poolChannel <- IDedPoolStatisticGetter{ID: id, Getter: pool}
		}
		close(poolChannel)
	}()
	return poolChannel
}

// gatherPoolStatistics will collect, in parallel, stats against each StoragePool referenced by 'pool'
func (s *PowerFlexService) gatherPoolStatistics(_ context.Context, scMeta *StorageClassMeta, pool <-chan IDedPoolStatisticGetter) <-chan *storagePoolMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherPoolStatistics")

	ch := make(chan *storagePoolMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerFlexConnections)

	go func() {
		for pl := range pool {
			wg.Add(1)
			sem <- struct{}{}
			go func(pl IDedPoolStatisticGetter) {
				defer wg.Done()
				defer func() {
					<-sem
				}()

				if pl.Getter.GetGen() == types.GenTypeEC {
					s.Logger.WithFields(logrus.Fields{
						"pool_id_used_for_metrics": pl.ID,
					}).Debug("calling GetMetrics(storage_pool)")

					stats, err := pl.Getter.GetClient().GetMetrics("storage_pool", []string{pl.ID})
					if err != nil {
						s.Logger.WithError(err).WithField("pool_id", pl.ID).Error("getting statistics pool")
						return
					}

					if len(stats.Resources) == 0 {
						s.Logger.Warn("No resources found in metrics response for storage pool")
						return
					}
					// convert bytes to GB
					storgaePoolMetrcisMap := stats.Resources[0].Metrics
					if storgaePoolMetrcisMap == nil {
						s.Logger.WithField("pool_id", pl.ID).Warn("metrics map is nil")
						return
					}

					const giB = float64(1 << 30)
					getMetric := func(metricName string) float64 {
						metricValue := getMetric(storgaePoolMetrcisMap, metricName)
						return metricValue / giB
					}

					totalCapacity := getMetric("physical_total")
					capacityAvailable := getMetric("physical_free")
					capacityInUse := getMetric("physical_used")
					provisioned := getMetric("logical_provisioned")

					s.Logger.WithFields(logrus.Fields{
						"pool_id":                    pl.ID,
						"storage_class_meta":         scMeta,
						"total_logical_capacity":     totalCapacity,
						"logical_capacity_available": capacityAvailable,
						"logical_capacity_in_use":    capacityInUse,
						"logical_provisioned":        provisioned,
					}).Debug("pool statistics")

					ch <- &storagePoolMetricsRecord{
						ID:                       pl.ID,
						storageClassMeta:         scMeta,
						TotalLogicalCapacity:     totalCapacity,
						LogicalCapacityAvailable: capacityAvailable,
						LogicalCapacityInUse:     capacityInUse,
						LogicalProvisioned:       provisioned,
					}
				} else {
					stats, err := pl.Getter.GetStatisticsGetter().GetStatistics()
					if err != nil {
						s.Logger.WithError(err).WithField("pool_id", pl.ID).Error("getting statistics pool")
						return
					}

					totalLogicalCapacity := GetTotalLogicalCapacity(stats)
					logicalCapacityAvailable := GetLogicalCapacityAvailable(stats)
					logicalCapacityInUse := GetLogicalCapacityInUse(stats)
					logicalProvisioned := GetLogicalProvisioned(stats)

					s.Logger.WithFields(logrus.Fields{
						"pool_id":                    pl.ID,
						"storage_class_meta":         scMeta,
						"total_logical_capacity":     totalLogicalCapacity,
						"logical_capacity_available": logicalCapacityAvailable,
						"logical_capacity_in_use":    logicalCapacityInUse,
						"logical_provisioned":        logicalProvisioned,
					}).Debug("pool statistics")

					ch <- &storagePoolMetricsRecord{
						ID:                       pl.ID,
						storageClassMeta:         scMeta,
						TotalLogicalCapacity:     totalLogicalCapacity,
						LogicalCapacityAvailable: logicalCapacityAvailable,
						LogicalCapacityInUse:     logicalCapacityInUse,
						LogicalProvisioned:       logicalProvisioned,
					}
				}
			}(pl)
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

// pushPoolStatistics will, in parallel, record stats in 'metrics' using the s.MetricsWrapper
func (s *PowerFlexService) pushPoolStatistics(ctx context.Context, spMetricRecord <-chan *storagePoolMetricsRecord) <-chan *storagePoolMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "pushPoolStatistics")
	var wg sync.WaitGroup

	ch := make(chan *storagePoolMetricsRecord)
	go func() {
		for i := range spMetricRecord {
			wg.Add(1)
			go func(i *storagePoolMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordCapacity(ctx, *(i.storageClassMeta), i.TotalLogicalCapacity, i.LogicalCapacityAvailable, i.LogicalCapacityInUse, i.LogicalProvisioned)
				if err != nil {
					s.Logger.WithError(err).Error("recording statistics for storage pool")
				}
				ch <- i
			}(i)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// ExportTopologyMetrics will export topology metrics
func (s *PowerFlexService) ExportTopologyMetrics(ctx context.Context) {
	start := time.Now()
	defer s.timeSince(start, "ExportTopologyMetrics")

	if s.MetricsWrapper == nil {
		s.Logger.Warn("no MetricsWrapper provided for getting ExportTopologyMetrics")
		return
	}

	pvs, err := s.VolumeFinder.GetPersistentVolumes()
	if err != nil {
		s.Logger.WithError(err).Error("getting persistent volumes")
		return
	}

	for range s.pushTopologyMetrics(ctx, s.gatherTopologyMetrics(s.getVolumeInfo(ctx, pvs))) {
		// consume the channel until it is empty and closed
	} // revive:disable-line:empty-block
}

// gatherTopologyMetrics will return a channel of topology metrics
func (s *PowerFlexService) gatherTopologyMetrics(volumes <-chan k8s.VolumeInfo) <-chan *TopologyMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherTopologyMetrics")

	ch := make(chan *TopologyMetricsRecord)
	var wg sync.WaitGroup

	go func() {
		for volume := range volumes {
			wg.Add(1)
			go func(volume k8s.VolumeInfo) {
				defer wg.Done()

				volumeProperties := strings.Split(volume.VolumeHandle, "-")
				if len(volumeProperties) != ExpectedVolumeHandleProperties {
					s.Logger.WithField("volume_handle", volume.VolumeHandle).Warn("unable to get VolumeID and ClusterID from volume handle")
					return
				}

				topologyMeta := &TopologyMeta{
					Namespace:               volume.Namespace,
					PersistentVolumeClaim:   volume.VolumeClaimName,
					VolumeClaimName:         volume.PersistentVolume,
					PersistentVolumeStatus:  volume.PersistentVolumeStatus,
					PersistentVolume:        volume.PersistentVolume,
					StorageClass:            volume.StorageClass,
					Driver:                  volume.Driver,
					ProvisionedSize:         volume.ProvisionedSize,
					StorageSystemVolumeName: volume.StorageSystemVolumeName,
					StoragePoolName:         volume.StoragePoolName,
					StorageSystem:           volume.StorageSystem,
					Protocol:                volume.Protocol,
					CreatedTime:             volume.CreatedTime,
				}

				pvAvailable := int64(1)

				metric := &TopologyMetricsRecord{
					topologyMeta: topologyMeta,
					pvAvailable:  pvAvailable,
				}

				ch <- metric
			}(volume)
		}

		wg.Wait()
		close(ch)
	}()
	return ch
}

// pushTopologyMetrics will push the provided channel of volume metrics to a data collector
func (s *PowerFlexService) pushTopologyMetrics(ctx context.Context, topologyMetrics <-chan *TopologyMetricsRecord) <-chan *TopologyMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "pushTopologyMetrics")
	var wg sync.WaitGroup

	ch := make(chan *TopologyMetricsRecord)
	go func() {
		for metrics := range topologyMetrics {
			wg.Add(1)
			go func(metrics *TopologyMetricsRecord) {
				defer wg.Done()
				err := s.MetricsWrapper.RecordTopologyMetrics(ctx, metrics.topologyMeta, metrics)
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", metrics.topologyMeta.PersistentVolume).Error("recording topology metrics for volume")
				} else {
					ch <- metrics
				}
			}(metrics)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// GetSDCBandwidth returns the read and write bandwidth based on the given SDC statistics
func GetSDCBandwidth(stats *types.SdcStatistics) (readBW float64, writeBW float64) {
	readBW = 0.0
	writeBW = 0.0

	if stats == nil {
		return readBW, writeBW
	}

	if stats.UserDataReadBwc.NumSeconds > 0 {
		readBW = float64(stats.UserDataReadBwc.TotalWeightInKb/stats.UserDataReadBwc.NumSeconds) / 1024.0
	}
	if stats.UserDataWriteBwc.NumSeconds > 0 {
		writeBW = float64(stats.UserDataWriteBwc.TotalWeightInKb/stats.UserDataWriteBwc.NumSeconds) / 1024.0
	}

	return readBW, writeBW
}

// GetSDCIOPS returns the read and write IOPS based on the given SDC statistics
func GetSDCIOPS(stats *types.SdcStatistics) (readIOPS float64, writeIOPS float64) {
	readIOPS = 0.0
	writeIOPS = 0.0

	if stats == nil {
		return readIOPS, writeIOPS
	}

	iopsCalculation := func(numOccurred, numSeconds int) float64 {
		return float64(numOccurred) / float64(numSeconds)
	}

	if stats.UserDataReadBwc.NumSeconds > 0 {
		readIOPS = iopsCalculation(stats.UserDataReadBwc.NumOccured, stats.UserDataReadBwc.NumSeconds)
	}

	if stats.UserDataWriteBwc.NumSeconds > 0 {
		writeIOPS = iopsCalculation(stats.UserDataWriteBwc.NumOccured, stats.UserDataWriteBwc.NumSeconds)
	}

	return readIOPS, writeIOPS
}

// GetSDCLatency returns the read and write latency based on the given SDC statistics
func GetSDCLatency(stats *types.SdcStatistics) (readLatency float64, writeLatency float64) {
	readLatency = 0.0
	writeLatency = 0.0
	if stats == nil {
		return readLatency, writeLatency
	}
	latencyCalculation := func(totalWeightInKb, numOccurred int) float64 {
		return float64(totalWeightInKb) / float64(numOccurred) / 1024.0
	}
	if stats.UserDataSdcReadLatency.NumOccured > 0 {
		readLatency = latencyCalculation(stats.UserDataSdcReadLatency.TotalWeightInKb, stats.UserDataSdcReadLatency.NumOccured)
	}
	if stats.UserDataSdcWriteLatency.NumOccured > 0 {
		writeLatency = latencyCalculation(stats.UserDataSdcWriteLatency.TotalWeightInKb, stats.UserDataSdcWriteLatency.NumOccured)
	}
	return readLatency, writeLatency
}

// GetVolumeBandwidth returns the read and write bandwidth based on the given SDC statistics
func GetVolumeBandwidth(stats *VolumeMetaMetrics) (readBW float64, writeBW float64) {
	readBW = 0.0
	writeBW = 0.0

	if stats == nil {
		return readBW, writeBW
	}
	if stats.GenType == types.GenTypeEC {
		readBW = stats.HostReadBandwith
		writeBW = stats.HostWriteBandwith
	} else {
		if stats.ReadBwc.NumSeconds > 0 {
			readBW = float64(stats.ReadBwc.TotalWeightInKb/stats.ReadBwc.NumSeconds) / 1024.0
		}
		if stats.WriteBwc.NumSeconds > 0 {
			writeBW = float64(stats.WriteBwc.TotalWeightInKb/stats.WriteBwc.NumSeconds) / 1024.0
		}
	}

	return readBW, writeBW
}

// GetVolumeIOPS returns the read and write IOPS based on the given SDC statistics
func GetVolumeIOPS(stats *VolumeMetaMetrics) (readIOPS float64, writeIOPS float64) {
	readIOPS = 0.0
	writeIOPS = 0.0

	if stats == nil {
		return readIOPS, writeIOPS
	}

	if stats.GenType == types.GenTypeEC {
		readIOPS = stats.HostReadIOPS
		writeIOPS = stats.HostWriteIOPS
	} else {
		iopsCalculation := func(numOccurred, numSeconds int) float64 {
			return float64(numOccurred) / float64(numSeconds)
		}

		if stats.ReadBwc.NumSeconds > 0 {
			readIOPS = iopsCalculation(stats.ReadBwc.NumOccured, stats.ReadBwc.NumSeconds)
		}

		if stats.WriteBwc.NumSeconds > 0 {
			writeIOPS = iopsCalculation(stats.WriteBwc.NumOccured, stats.WriteBwc.NumSeconds)
		}
	}
	return readIOPS, writeIOPS
}

// GetVolumeLatency returns the read and write latency based on the given SDC statistics
func GetVolumeLatency(stats *VolumeMetaMetrics) (readLatency float64, writeLatency float64) {
	readLatency = 0.0
	writeLatency = 0.0
	if stats == nil {
		return readLatency, writeLatency
	}
	if stats.GenType == types.GenTypeEC {
		readLatency = stats.AvgHostReadLatency
		writeLatency = stats.AvgHostWriteLatency
	} else {
		latencyCalculation := func(totalWeightInKb, numOccurred int) float64 {
			return float64(totalWeightInKb) / float64(numOccurred) / 1024.0
		}
		if stats.ReadLatencyBwc.NumOccured > 0 {
			readLatency = latencyCalculation(stats.ReadLatencyBwc.TotalWeightInKb, stats.ReadLatencyBwc.NumOccured)
		}
		if stats.WriteLatencyBwc.NumOccured > 0 {
			writeLatency = latencyCalculation(stats.WriteLatencyBwc.TotalWeightInKb, stats.WriteLatencyBwc.NumOccured)
		}
	}
	return readLatency, writeLatency
}

// GetTotalLogicalCapacity returns the used + unused user data in GB from the given storage pool statistics
func GetTotalLogicalCapacity(stats *types.Statistics) float64 {
	if stats == nil {
		return 0
	}

	return (float64(stats.NetUnusedCapacityInKb) + float64(stats.NetUserDataCapacityInKb)) / (1024.0 * 1024.0)
}

// GetLogicalCapacityAvailable returns the unused user data in GB from the given storage pool statistics
func GetLogicalCapacityAvailable(stats *types.Statistics) float64 {
	if stats == nil {
		return 0
	}

	return float64(stats.NetUnusedCapacityInKb) / (1024.0 * 1024.0)
}

// GetLogicalCapacityInUse returns the used user data in GB from the given storage pool statistics
func GetLogicalCapacityInUse(stats *types.Statistics) float64 {
	if stats == nil {
		return 0
	}

	return float64(stats.NetUserDataCapacityInKb) / (1024.0 * 1024.0)
}

// GetLogicalProvisioned returns the total volume size in GB from the given storage pool statistics
func GetLogicalProvisioned(stats *types.Statistics) float64 {
	if stats == nil {
		return 0
	}

	return float64(stats.VolumeAddressSpaceInKb) / (1024.0 * 1024.0)
}

// contains checks if a string slice contains a specific string value
func contains(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

// timeSince logs the duration since start time with the given function name
func (s *PowerFlexService) timeSince(start time.Time, fName string) {
	s.Logger.WithFields(logrus.Fields{
		"duration": fmt.Sprintf("%v", time.Since(start)),
		"function": fName,
	}).Info("function duration")
}
