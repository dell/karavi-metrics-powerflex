// Copyright © 2025 Dell Inc. or its subsidiaries. All Rights Reserved.
//
// Dell Technologies, Dell and other trademarks are trademarks of Dell Inc.
// or its subsidiaries. Other trademarks may be trademarks of their respective
// owners.

package domain

// ArrayConnectionData contains data required to connect to array
type ArrayConnectionData struct {
	SystemID                  string            `json:"systemID"`
	Username                  string            `json:"username"`
	Password                  string            `json:"password"`
	Endpoint                  string            `json:"endpoint"`
	Insecure                  bool              `json:"insecure,omitempty"`
	IsDefault                 bool              `json:"isDefault,omitempty"`
	SkipCertificateValidation bool              `json:"skipCertificateValidation,omitempty"`
	AvailabilityZone          *AvailabilityZone `json:"zone,omitempty"`
}

// Definitions to make AvailabilityZone decomposition easier to read.
type (
	ZoneName             string
	ProtectionDomainName string
	PoolName             string
)

// AvailabilityZone provides a mapping between cluster zones labels and storage systems
type AvailabilityZone struct {
	Name              ZoneName           `json:"name"`
	LabelKey          string             `json:"labelKey"`
	ProtectionDomains []ProtectionDomain `json:"protectionDomains"`
}

// ProtectionDomain provides protection domain information for a cluster's availability zone
type ProtectionDomain struct {
	Name  ProtectionDomainName `json:"name"`
	Pools []PoolName           `json:"pools"`
}
