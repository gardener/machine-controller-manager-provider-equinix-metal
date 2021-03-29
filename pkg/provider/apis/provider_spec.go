/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
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

package api

const (
	// APIKey is a constant for a key name that is part of the Packet cloud credentials
	APIKey string = "apiToken"
)

// EquinixMetalProviderSpec is the spec to be used while parsing the calls.
type EquinixMetalProviderSpec struct {
	APIVersion   string   `json:"apiVersion,omitempty"`
	Facility     []string `json:"facility"`
	MachineType  string   `json:"machineType"`
	BillingCycle string   `json:"billingCycle"`
	OS           string   `json:"OS"`
	ProjectID    string   `json:"projectID"`
	Tags         []string `json:"tags,omitempty"`
	SSHKeys      []string `json:"sshKeys,omitempty"`
	UserData     string   `json:"userdata,omitempty"`
}
