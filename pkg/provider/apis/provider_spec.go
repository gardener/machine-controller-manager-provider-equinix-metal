// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package api

const (
	// APIKey is a constant for a key name that is part of the equinix metal cloud credentials
	APIKey string = "apiToken"
	// AlternateAPIKey is an alternate constant for a key name that is part of the equinix metal cloud credentials.
	// It is used as an alternative when APIKey isn't found.
	AlternateAPIKey string = "alternateApiToken"
	// V1alpha1 is the API version
	V1alpha1 string = "mcm.gardener.cloud/v1alpha1"
)

// EquinixMetalProviderSpec is the spec to be used while parsing the calls.
type EquinixMetalProviderSpec struct {
	APIVersion     string   `json:"apiVersion,omitempty"`
	Metro          string   `json:"metro,omitempty"`
	MachineType    string   `json:"machineType"`
	BillingCycle   string   `json:"billingCycle"`
	OS             string   `json:"OS,omitempty"`
	IPXEScriptURL  *string  `json:"ipxeScriptUrl,omitempty"`
	ProjectID      string   `json:"projectID"`
	Tags           []string `json:"tags,omitempty"`
	SSHKeys        []string `json:"sshKeys,omitempty"`
	UserData       string   `json:"userdata,omitempty"`
	ReservationIDs []string `json:"reservationIDs,omitempty"`
	ReservedOnly   bool     `json:"reservedDevicesOnly,omitempty"`
}
