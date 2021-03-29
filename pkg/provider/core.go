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

// Package provider contains the cloud provider specific implementations to manage machines
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	validation "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis/validation"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/packethost/packngo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
)

// Driver is the driver struct for holding EquinixMetal machine information
type Driver struct {
}

// NOTE
//
// The basic working of the controller will work with just implementing the CreateMachine() & DeleteMachine() methods.
// You can first implement these two methods and check the working of the controller.
// Leaving the other methods to NOT_IMPLEMENTED error status.
// Once this works you can implement the rest of the methods.
//
// Also make sure each method return appropriate errors mentioned in `https://github.com/gardener/machine-controller-manager/blob/master/docs/development/machine_error_codes.md`

// CreateMachine handles a machine creation request
// REQUIRED METHOD
//
// REQUEST PARAMETERS (driver.CreateMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be created
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.CreateMachineResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                                ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                                Eg: gce://project-name/region/vm-ProviderID
// NodeName              string                   Returns the name of the node-object that the VM register's with Kubernetes.
//                                                This could be different from req.MachineName as well
// LastKnownState        string                   (Optional) Last known state of VM during the current operation.
//                                                Could be helpful to continue operations in future requests.
//
// OPTIONAL IMPLEMENTATION LOGIC
// It is optionally expected by the safety controller to use an identification mechanisms to map the VM Created by a providerSpec.
// These could be done using tag(s)/resource-groups etc.
// This logic is used by safety controller to delete orphan VMs which are not backed by any machine CRD
//
func (p *Provider) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track request
	klog.V(2).Infof("Machine creation request has been received for %q", req.Machine.Name)

	var (
		exists       bool
		userData     []byte
		machine      = req.Machine
		secret       = req.Secret
		machineClass = req.MachineClass
	)

	providerSpec, err := decodeProviderSpecAndSecret(machineClass, secret)
	if err != nil {
		return nil, err
	}

	svc := p.createSVC(getApiKey(secret))
	if svc == nil {
		return nil, fmt.Errorf("nil Equinix Metal service returned")
	}
	if userData, exists = secret.Data["userData"]; !exists {
		return nil, status.Error(codes.Internal, "userData doesn't exist")
	}
	// packet tags are strings only
	createRequest := &packngo.DeviceCreateRequest{
		Hostname:       machine.Name,
		UserData:       string(userData),
		Plan:           providerSpec.MachineType,
		ProjectID:      providerSpec.ProjectID,
		BillingCycle:   providerSpec.BillingCycle,
		Facility:       providerSpec.Facility,
		OS:             providerSpec.OS,
		ProjectSSHKeys: providerSpec.SSHKeys,
		Tags:           providerSpec.Tags,
	}

	device, _, err := svc.Devices.Create(createRequest)
	if err != nil {
		klog.Errorf("Could not create machine: %v", err)
		return nil, err
	}
	response := &driver.CreateMachineResponse{
		ProviderID: device.ID,
		NodeName:   machine.Name,
	}
	klog.V(2).Infof("Machine creation request has been processed for %q", machine.Name)

	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (driver.DeleteMachineRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM is to be deleted
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.DeleteMachineResponse)
// LastKnownState        bytes(blob)              (Optional) Last known state of VM during the current operation.
//                                                Could be helpful to continue operations in future requests.
//
func (p *Provider) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", req.Machine.Name)

	instanceID := decodeMachineID(req.Machine.Spec.ProviderID)
	svc := p.createSVC(getApiKey(req.Secret))
	if svc == nil {
		return nil, fmt.Errorf("nil Packet service returned")
	}
	resp, err := svc.Devices.Delete(instanceID)
	if err != nil {
		if resp.StatusCode == 404 {
			// if it is not found, do not error, just return
			klog.V(2).Infof("No machine matching the machine-ID found on the provider %q", instanceID)
			return &driver.DeleteMachineResponse{}, nil
		}
		klog.Errorf("Could not terminate machine %s: %v", instanceID, err)
		return nil, fmt.Errorf("Could not terminate machine %s: %v", instanceID, err)
	}
	klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)
	return &driver.DeleteMachineResponse{}, nil
}

// us handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.GetMachineStatusRequest)
// Machine               *v1alpha1.Machine        Machine object from whom VM status needs to be returned
// MachineClass          *v1alpha1.MachineClass   MachineClass backing the machine object
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.GetMachineStatueResponse)
// ProviderID            string                   Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                                ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                                Eg: gce://project-name/region/vm-ProviderID
// NodeName             string                    Returns the name of the node-object that the VM register's with Kubernetes.
//                                                This could be different from req.MachineName as well
//
// The request should return a NOT_FOUND (5) status error code if the machine is not existing
func (p *Provider) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("Get request has been received for %q", req.Machine.Name)

	var (
		id   = req.Machine.Spec.ProviderID
		name = req.Machine.Name
	)
	svc := p.createSVC(getApiKey(req.Secret))
	if svc == nil {
		return nil, fmt.Errorf("nil Packet service returned")
	}
	device, _, err := svc.Devices.Get(id, &packngo.GetOptions{})
	if err != nil {
		klog.Errorf("Could not get device %s: %v", id, err)
		return nil, err
	}

	klog.V(2).Infof("Machine get request has been processed successfully for %q", name)
	return &driver.GetMachineStatusResponse{
		NodeName:   name,
		ProviderID: encodeMachineID(device.Facility.String(), device.ID),
	}, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// MachineClass          *v1alpha1.MachineClass   MachineClass based on which VMs created have to be listed
// Secret                *corev1.Secret           Kubernetes secret that contains any sensitive data/credentials
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//                                           for all machine's who where possibilly created by this ProviderSpec
//
func (p *Provider) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("List machines request has been received for %q", req.MachineClass.Name)

	var (
		clusterName, nodeRole string
		resp                  = &driver.ListMachinesResponse{
			MachineList: make(map[string]string),
		}
	)

	providerSpec, err := decodeProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, err
	}

	for _, key := range providerSpec.Tags {
		if strings.Contains(key, "kubernetes.io/cluster/") {
			clusterName = key
		} else if strings.Contains(key, "kubernetes.io/role/") {
			nodeRole = key
		}
	}

	if clusterName == "" || nodeRole == "" {
		return resp, nil
	}

	svc := p.createSVC(getApiKey(req.Secret))
	if svc == nil {
		return nil, fmt.Errorf("nil Equinix Metal service returned")
	}
	devices, _, err := svc.Devices.List(providerSpec.ProjectID, &packngo.ListOptions{})
	if err != nil {
		klog.Errorf("Could not list devices for project %s: %v", providerSpec.ProjectID, err)
		return nil, err
	}
	for _, d := range devices {
		matchedCluster := false
		matchedRole := false
		for _, tag := range d.Tags {
			switch tag {
			case clusterName:
				matchedCluster = true
			case nodeRole:
				matchedRole = true
			}
		}
		if matchedCluster && matchedRole {
			resp.MachineList[d.ID] = d.Hostname
		}
	}
	return resp, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (driver.GetVolumeIDsRequest)
// PVSpecList            []*corev1.PersistentVolumeSpec       PVSpecsList is a list PV specs for whom volume-IDs are required.
//
// RESPONSE PARAMETERS (driver.GetVolumeIDsResponse)
// VolumeIDs             []string                             VolumeIDs is a repeated list of VolumeIDs.
//
func (p *Provider) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecs)
	defer klog.V(2).Infof("GetVolumeIDs request has been processed successfully for %q", req.PVSpecs)

	return &driver.GetVolumeIDsResponse{}, status.Error(codes.Unimplemented, "Equinix Metal does not have storage")
}

// GenerateMachineClassForMigration helps in migration of one kind of machineClass CR to another kind.
// For instance an machineClass custom resource of `AWSMachineClass` to `MachineClass`.
// Implement this functionality only if something like this is desired in your setup.
// If you don't require this functionality leave is as is. (return Unimplemented)
//
// The following are the tasks typically expected out of this method
// 1. Validate if the incoming classSpec is valid one for migration (e.g. has the right kind).
// 2. Migrate/Copy over all the fields/spec from req.ProviderSpecificMachineClass to req.MachineClass
// For an example refer
//		https://github.com/prashanth26/machine-controller-manager-provider-gcp/blob/migration/pkg/gcp/machine_controller.go#L222-L233
//
// REQUEST PARAMETERS (driver.GenerateMachineClassForMigration)
// ProviderSpecificMachineClass    interface{}                             ProviderSpecificMachineClass is provider specfic machine class object (E.g. AWSMachineClass). Typecasting is required here.
// MachineClass 				   *v1alpha1.MachineClass                  MachineClass is the machine class object that is to be filled up by this method.
// ClassSpec                       *v1alpha1.ClassSpec                     Somemore classSpec details useful while migration.
//
// RESPONSE PARAMETERS (driver.GenerateMachineClassForMigration)
// NONE
//
func (p *Provider) GenerateMachineClassForMigration(ctx context.Context, req *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	// Log messages to track start and end of request
	klog.V(2).Infof("MigrateMachineClass request has been received for %q", req.ClassSpec)
	defer klog.V(2).Infof("MigrateMachineClass request has been processed successfully for %q", req.ClassSpec)

	return &driver.GenerateMachineClassForMigrationResponse{}, status.Error(codes.Unimplemented, "")
}

//  create a session
func (p *Provider) createSVC(apiKey string) *packngo.Client {

	token := strings.TrimSpace(apiKey)

	if token != "" {
		return packngo.NewClientWithAuth("gardener", token, nil)
	}

	return nil
}

// decodeProviderSpecAndSecret converts request parameters to api.ProviderSpec & api.Secrets
func decodeProviderSpecAndSecret(machineClass *v1alpha1.MachineClass, secret *corev1.Secret) (*api.EquinixMetalProviderSpec, error) {
	var (
		providerSpec *api.EquinixMetalProviderSpec
	)

	// Extract providerSpec
	if machineClass == nil {
		return nil, status.Error(codes.Internal, "MachineClass ProviderSpec is nil")
	}

	err := json.Unmarshal(machineClass.ProviderSpec.Raw, &providerSpec)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Validate the Spec and Secrets
	validationErr := validation.ValidateProviderSpecNSecret(providerSpec, secret, field.NewPath("providerSpec"))
	if validationErr.ToAggregate() != nil && len(validationErr.ToAggregate().Errors()) > 0 {
		err = fmt.Errorf("Error while validating ProviderSpec %v", validationErr.ToAggregate().Error())
		klog.V(2).Infof("Validation of EquinixMetalMachineClass failed %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}

func encodeMachineID(facility, machineID string) string {
	return fmt.Sprintf("equinixmetal:///%s/%s", facility, machineID)
}

func decodeMachineID(id string) string {
	splitProviderID := strings.Split(id, "/")
	return splitProviderID[len(splitProviderID)-1]
}
