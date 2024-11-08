/*
 Copyright (c) 2020-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

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
	"sync"
	"time"

	"github.com/dell/karavi-metrics-powerflex/internal/k8s"
	"github.com/sirupsen/logrus"

	sio "github.com/dell/goscaleio"
	types "github.com/dell/goscaleio/types/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/storage/v1"
)

var _ Service = (*PowerFlexService)(nil)

const (
	// DefaultMaxPowerFlexConnections is the number of workers that can query powerflex  at a time
	DefaultMaxPowerFlexConnections = 10
)

// Service contains operations that would be used to interact with a PowerFlex system
//
//go:generate mockgen -destination=mocks/service_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service Service
type Service interface {
	GetSDCs(context.Context, PowerFlexClient, SDCFinder) ([]StatisticsGetter, error)
	GetSDCStatistics(context.Context, []corev1.Node, []StatisticsGetter)
	GetVolumes(context.Context, []StatisticsGetter) ([]VolumeStatisticsGetter, error)
	ExportVolumeStatistics(context.Context, []VolumeStatisticsGetter, VolumeFinder)
	GetStorageClasses(ctx context.Context, client PowerFlexClient, storageClassFinder StorageClassFinder) ([]StorageClassMeta, error)
	GetStoragePoolStatistics(ctx context.Context, storageClassMetas []StorageClassMeta)
}

// StatisticsGetter supports getting statistics
//
//go:generate mockgen -destination=mocks/statistics_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service StatisticsGetter
type StatisticsGetter interface {
	GetStatistics() (*types.SdcStatistics, error)
	GetVolume() ([]*types.Volume, error)
	FindVolumes() ([]*sio.Volume, error)
}

// VolumeStatisticsGetter supports getting statistics
//
//go:generate mockgen -destination=mocks/volume_statistics_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service VolumeStatisticsGetter
type VolumeStatisticsGetter interface {
	GetVolumeStatistics() (*types.VolumeStatistics, error)
}

// TokenGetter gets a powerflex login token
//
//go:generate mockgen -destination=mocks/token_getter_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service TokenGetter
type TokenGetter interface {
	GetToken(ctx context.Context) (string, error)
	Stop()
}

// PowerFlexClient contains operations for a powerflex client
//
//go:generate mockgen -destination=mocks/powerflex_client_mocks.go -package=mocks github.com/dell/karavi-metrics-powerflex/internal/service PowerFlexClient
type PowerFlexClient interface {
	//Authenticate(*sio.ConfigConnect) (sio.Cluster, error)
	GetInstance(string) ([]*types.System, error)
	FindSystem(string, string, string) (*sio.System, error)
	GetStoragePool(href string) ([]*types.StoragePool, error)
	GetConfigConnect() *sio.ConfigConnect
	SetToken(string)
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
	GetStorageClasses() ([]v1.StorageClass, error)
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

// GetSDCs returns a slice of SDCs
func (s *PowerFlexService) GetSDCs(_ context.Context, client PowerFlexClient, sdcFinder SDCFinder) ([]StatisticsGetter, error) {
	var sdcs []StatisticsGetter
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
		s.Logger.WithFields(logrus.Fields{"system": sys}).Debug("found system")
		for _, sdcGUID := range sdcGUIDs {
			sdc, err := sys.FindSdc("SdcGUID", sdcGUID)
			if err != nil {
				s.Logger.WithField("sdc_guid", sdcGUID).Warn("unable to find SDC with GUID")
			} else {
				s.Logger.WithFields(logrus.Fields{"sdc_guid": sdcGUID}).Debug("found sdc")
				sdcs = append(sdcs, sdc)
			}
		}
	}
	return sdcs, nil
}

// SystemFinder is a function that will be used for finding a PowerFlexSystem by id, name, and href
var SystemFinder = func(client PowerFlexClient, id string, name string, href string) (PowerFlexSystem, error) {
	return client.FindSystem(id, name, href)
}

// GetSDCMeta returns SDC meta information from a goscaleio SDC
// This function is exported for direct testing
func GetSDCMeta(sdc interface{}, nodes []corev1.Node) *SDCMeta {
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
		}
	default:
		return &SDCMeta{}
	}
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

// GetSDCStatistics records I/O statistics for the given list of SDCs
func (s *PowerFlexService) GetSDCStatistics(ctx context.Context, nodes []corev1.Node, sdcs []StatisticsGetter) {
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

func (s *PowerFlexService) sdcServer(sdcs []StatisticsGetter) <-chan StatisticsGetter {
	sdcChan := make(chan StatisticsGetter, len(sdcs))
	go func() {
		for _, sdc := range sdcs {
			sdcChan <- sdc
		}
		close(sdcChan)
	}()
	return sdcChan
}

// gatherSDCMetrics will collect, in parallel, stats against each SDC referenced by 'statGetters'
func (s *PowerFlexService) gatherSDCMetrics(_ context.Context, nodes []corev1.Node, sdcs <-chan StatisticsGetter) <-chan *SDCMetricsRecord {
	start := time.Now()
	defer s.timeSince(start, "gatherMetrics")

	ch := make(chan *SDCMetricsRecord)
	var wg sync.WaitGroup
	sem := make(chan struct{}, s.MaxPowerFlexConnections)

	go func() {
		for sdc := range sdcs {
			wg.Add(1)
			sem <- struct{}{}
			go func(sdc StatisticsGetter) {
				defer func() {
					wg.Done()
					<-sem
				}()

				sdcMeta := GetSDCMeta(sdc, nodes)

				stats, err := sdc.GetStatistics()
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
					ch <- fmt.Sprintf(mr.sdcMeta.ID)
				}
			}(record)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

func getVolumeMeta(volume interface{}) *VolumeMeta {
	switch v := volume.(type) {
	case *sio.Volume:
		var sdcsInfo []MappedSDC
		for _, d := range v.Volume.MappedSdcInfo {
			sdcsInfo = append(sdcsInfo,
				MappedSDC{SdcID: d.SdcID, SdcIP: d.SdcIP},
			)
		}
		return &VolumeMeta{
			Name:       v.Volume.Name,
			ID:         v.Volume.ID,
			MappedSDCs: sdcsInfo,
		}
	default:
		return &VolumeMeta{
			Name:       "",
			ID:         "",
			MappedSDCs: []MappedSDC{},
		}
	}
}

// GetVolumes returns all unique, mapped volumes in sdcs
func (s *PowerFlexService) GetVolumes(_ context.Context, sdcs []StatisticsGetter) ([]VolumeStatisticsGetter, error) {
	var uniqueVolumes []VolumeStatisticsGetter
	visited := make(map[string]bool)

	for _, sdc := range sdcs {
		vols, err := sdc.FindVolumes()
		if err != nil {
			return nil, err
		}
		for _, v := range vols {
			volumeMeta := getVolumeMeta(v)
			if !visited[volumeMeta.ID] {
				s.Logger.WithField("volume_id", volumeMeta.ID).Debug("found volume")
				uniqueVolumes = append(uniqueVolumes, v)
				visited[volumeMeta.ID] = true
			}
		}
	}

	return uniqueVolumes, nil
}

// ExportVolumeStatistics records I/O statistics for the given list of Volumes
func (s *PowerFlexService) ExportVolumeStatistics(ctx context.Context, volumes []VolumeStatisticsGetter, volumeFinder VolumeFinder) {
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
func (s *PowerFlexService) volumeServer(volumes []VolumeStatisticsGetter) <-chan VolumeStatisticsGetter {
	volumeChannel := make(chan VolumeStatisticsGetter, len(volumes))
	go func() {
		for _, volume := range volumes {
			volumeChannel <- volume
		}
		close(volumeChannel)
	}()
	return volumeChannel
}

// gatherVolumeMetrics will return a channel of volume metrics based on the input of volumes
func (s *PowerFlexService) gatherVolumeMetrics(_ context.Context, volumeFinder VolumeFinder, volumes <-chan VolumeStatisticsGetter) <-chan *VolumeMetricsRecord {
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
			go func(volume VolumeStatisticsGetter) {
				defer func() {
					wg.Done()
					<-sem
				}()
				volumeMeta := getVolumeMeta(volume)
				if pv, ok := persistentVolumes[volumeMeta.Name]; ok {
					volumeMeta.PersistentVolumeName = pv.PersistentVolume
					volumeMeta.StorageSystemID = pv.StorageSystemID
					volumeMeta.Namespace = pv.Namespace
					volumeMeta.PersistentVolumeClaimName = pv.VolumeClaimName
				} else {
					s.Logger.WithField("volume_id", volumeMeta.ID).Error("could not find a Persistent Volume that maps to storage system volume ID")
				}

				stats, err := volume.GetVolumeStatistics()
				if err != nil {
					s.Logger.WithError(err).WithField("volume_id", volumeMeta.ID).Error("getting statistics for volume")
					return
				}
				readBW, writeBW := GetVolumeBandwidth(stats)
				readIOPS, writeIOPS := GetVolumeIOPS(stats)
				readLatency, writeLatency := GetVolumeLatency(stats)

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
					ch <- fmt.Sprintf(metrics.volumeMeta.ID)
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
			systemid := class.Parameters["systemID"]
			s.Logger.WithField("class.Parameters", systemid).Debug("systems log")
			if system.ID == systemid {
				storageClassInfo := StorageClassInfo{
					ID:              string(class.UID),
					Name:            class.Name,
					Driver:          class.Provisioner,
					StorageSystemID: systemid,
					StoragePools:    k8s.GetStoragePools(class),
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
			StoragePools:    make(map[string]StoragePoolStatisticsGetter),
		}

		for _, systemPool := range systemStoragePools {
			if contains(class.StoragePools, systemPool.Name) {
				storageClassMeta.StoragePools[systemPool.Name] = sio.NewStoragePoolEx(c, systemPool)
			}
		}

		storageClassMetas = append(storageClassMetas, storageClassMeta)
	}

	return storageClassMetas, nil
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
	Getter StoragePoolStatisticsGetter
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

func (s *PowerFlexService) storagePoolServer(pools map[string]StoragePoolStatisticsGetter) <-chan IDedPoolStatisticGetter {
	poolChannel := make(chan IDedPoolStatisticGetter)
	go func() {
		for id, pool := range pools {
			poolChannel <- IDedPoolStatisticGetter{ID: id, Getter: pool}
		}
		close(poolChannel)
	}()
	return poolChannel
}

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
				stats, err := pl.Getter.GetStatistics()
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
			}(pl)
		}
		wg.Wait()
		close(ch)
		close(sem)
	}()
	return ch
}

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

// GetSDCBandwidth returns the read and write bandwidth based on the given SDC statistics
func GetSDCBandwidth(stats *types.SdcStatistics) (readBW float64, writeBW float64) {
	readBW = 0.0
	writeBW = 0.0

	if stats == nil {
		return
	}

	if stats.UserDataReadBwc.NumSeconds > 0 {
		readBW = float64(stats.UserDataReadBwc.TotalWeightInKb/stats.UserDataReadBwc.NumSeconds) / 1024.0
	}
	if stats.UserDataWriteBwc.NumSeconds > 0 {
		writeBW = float64(stats.UserDataWriteBwc.TotalWeightInKb/stats.UserDataWriteBwc.NumSeconds) / 1024.0
	}

	return
}

// GetSDCIOPS returns the read and write IOPS based on the given SDC statistics
func GetSDCIOPS(stats *types.SdcStatistics) (readIOPS float64, writeIOPS float64) {
	readIOPS = 0.0
	writeIOPS = 0.0

	if stats == nil {
		return
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

	return
}

// GetSDCLatency returns the read and write latency based on the given SDC statistics
func GetSDCLatency(stats *types.SdcStatistics) (readLatency float64, writeLatency float64) {
	readLatency = 0.0
	writeLatency = 0.0
	if stats == nil {
		return
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
	return
}

// GetVolumeBandwidth returns the read and write bandwidth based on the given SDC statistics
func GetVolumeBandwidth(stats *types.VolumeStatistics) (readBW float64, writeBW float64) {
	readBW = 0.0
	writeBW = 0.0

	if stats == nil {
		return
	}

	if stats.UserDataReadBwc.NumSeconds > 0 {
		readBW = float64(stats.UserDataReadBwc.TotalWeightInKb/stats.UserDataReadBwc.NumSeconds) / 1024.0
	}
	if stats.UserDataWriteBwc.NumSeconds > 0 {
		writeBW = float64(stats.UserDataWriteBwc.TotalWeightInKb/stats.UserDataWriteBwc.NumSeconds) / 1024.0
	}

	return
}

// GetVolumeIOPS returns the read and write IOPS based on the given SDC statistics
func GetVolumeIOPS(stats *types.VolumeStatistics) (readIOPS float64, writeIOPS float64) {
	readIOPS = 0.0
	writeIOPS = 0.0

	if stats == nil {
		return
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

	return
}

// GetVolumeLatency returns the read and write latency based on the given SDC statistics
func GetVolumeLatency(stats *types.VolumeStatistics) (readLatency float64, writeLatency float64) {
	readLatency = 0.0
	writeLatency = 0.0
	if stats == nil {
		return
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
	return
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

func contains(slice []string, value string) bool {
	for _, element := range slice {
		if element == value {
			return true
		}
	}
	return false
}

func (s *PowerFlexService) timeSince(start time.Time, fName string) {
	s.Logger.WithFields(logrus.Fields{
		"duration": fmt.Sprintf("%v", time.Since(start)),
		"function": fName,
	}).Info("function duration")
}
