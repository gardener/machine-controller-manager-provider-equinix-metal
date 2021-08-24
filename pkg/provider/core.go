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
	"errors"
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

const (
	// PacketMachineClassKind is the deprecated CRD kind that was used for packet machine classes
	PacketMachineClassKind = "PacketMachineClass"
	// ProviderEquinixMetal is the provider type used to identify EquinixMetal
	ProviderEquinixMetal = "EquinixMetal"
)

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
		userData     []byte
		machine      = req.Machine
		secret       = req.Secret
		machineClass = req.MachineClass
	)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderEquinixMetal {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderEquinixMetal))
	}

	// decodes the provider spec, and validates the spec and the secret for required fields.
	providerSpec, err := decodeProviderSpec(machineClass)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validateSecret(secret); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())

	}
	// check that the name was valid
	if err := validation.ValidateName(machine.Name); len(err) > 0 {
		var msgs []string
		for _, e := range err {
			msgs = append(msgs, e.Error())
		}

		return nil, status.Error(codes.InvalidArgument, strings.Join(msgs, "; "))
	}

	svc := p.createSVC(req.Secret)
	if svc == nil {
		return nil, status.Error(codes.Internal, "nil Equinix Metal service returned")
	}
	// we already validated the existence and non-nil-ness of userData in the validation
	userData = secret.Data["userData"]

	// packet tags are strings only
	createRequest := &packngo.DeviceCreateRequest{
		Hostname:       machine.Name,
		UserData:       string(userData),
		Plan:           providerSpec.MachineType,
		ProjectID:      providerSpec.ProjectID,
		BillingCycle:   providerSpec.BillingCycle,
		Metro:          providerSpec.Metro,
		Facility:       providerSpec.Facility,
		OS:             providerSpec.OS,
		ProjectSSHKeys: providerSpec.SSHKeys,
		Tags:           providerSpec.Tags,
	}
	device, err := createDeviceWithReservations(svc, createRequest, providerSpec.ReservationID, providerSpec.ReservedOnly)
	if err != nil {
		klog.Errorf("Could not create machine: %v", err)
		return nil, status.Error(codes.Unavailable, fmt.Sprintf("Could not create machine: %v", err))
	}

	response := &driver.CreateMachineResponse{
		ProviderID: encodeMachineID(device),
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

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderEquinixMetal {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderEquinixMetal))
	}

	// decodes the provider spec, and validates the spec and the secret for required fields.
	if err := validateSecretAPIKey(req.Secret); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	instanceID := decodeMachineID(req.Machine.Spec.ProviderID)
	svc := p.createSVC(req.Secret)
	if svc == nil {
		return nil, status.Error(codes.Internal, "nil Equinix Metal service returned")
	}
	resp, err := svc.Delete(instanceID, true)
	if err != nil {
		if resp.StatusCode == 404 {
			// if it is not found, do not error, just return
			klog.V(2).Infof("No machine matching the machine-ID found on the provider %q", instanceID)
			return &driver.DeleteMachineResponse{}, nil
		}
		klog.Errorf("Could not terminate machine %s: %v", instanceID, err)
		return nil, status.Error(codes.Unknown, fmt.Sprintf("Could not terminate machine %s: %v", instanceID, err))
	}
	klog.V(2).Infof("Machine deletion request has been processed for %q", req.Machine.Name)
	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
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
		id   = decodeMachineID(req.Machine.Spec.ProviderID)
		name = req.Machine.Name
	)

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderEquinixMetal {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderEquinixMetal))
	}

	if err := validateSecretAPIKey(req.Secret); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	svc := p.createSVC(req.Secret)
	if svc == nil {
		return nil, status.Error(codes.Internal, "nil Equinix Metal service returned")
	}
	device, _, err := svc.Get(id, &packngo.GetOptions{})
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Could not get device %s: %v", id, err))
	}

	klog.V(2).Infof("Machine get request has been processed successfully for %q", name)
	return &driver.GetMachineStatusResponse{
		NodeName:   name,
		ProviderID: encodeMachineID(device),
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

	// Check if incoming CR is a CR we support
	if req.MachineClass.Provider != ProviderEquinixMetal {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderEquinixMetal))
	}

	providerSpec, err := decodeProviderSpec(req.MachineClass)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := validateSecretAPIKey(req.Secret); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
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

	svc := p.createSVC(req.Secret)
	if svc == nil {
		return nil, status.Error(codes.Internal, "nil Equinix Metal service returned")
	}
	devices, _, err := svc.List(providerSpec.ProjectID, &packngo.ListOptions{})
	if err != nil {
		msg := fmt.Sprintf("Could not list devices for project %s: %v", providerSpec.ProjectID, err)
		klog.Error(msg)
		return nil, status.Error(codes.Unknown, msg)
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
			resp.MachineList[encodeMachineID(&d)] = d.Hostname
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

	// this is the old PacketMachineClass; in the move to out-of-tree, we migrated to the newer Equinix Metal
	packetMachineClass := req.ProviderSpecificMachineClass.(*v1alpha1.PacketMachineClass)

	// Check if incoming CR is valid CR for migration
	// In this case, the MachineClassKind to be matching
	if req.ClassSpec.Kind != PacketMachineClassKind {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Requested for Provider '%s', we only support '%s'", req.MachineClass.Provider, ProviderEquinixMetal))
	}

	return &driver.GenerateMachineClassForMigrationResponse{}, fillUpMachineClass(packetMachineClass, req.MachineClass)
}

//  create a session
func (p *Provider) createSVC(secret *corev1.Secret) packngo.DeviceService {
	return p.SPI.NewSession(secret)
}

// decodeProviderSpec converts request parameters to api.ProviderSpec & api.Secrets
func decodeProviderSpec(machineClass *v1alpha1.MachineClass) (*api.EquinixMetalProviderSpec, error) {
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
	validationErr := validation.ValidateProviderSpec(providerSpec, field.NewPath("providerSpec"))
	if validationErr.ToAggregate() != nil && len(validationErr.ToAggregate().Errors()) > 0 {
		err = fmt.Errorf("Error while validating ProviderSpec %v", validationErr.ToAggregate().Error())
		klog.V(2).Infof("Validation of EquinixMetalMachineClass failed %s", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}

func createDeviceWithReservations(svc packngo.DeviceService, createRequest *packngo.DeviceCreateRequest, reservationIDs []string, reservedOnly bool) (device *packngo.Device, err error) {
	// if there were no reservation IDs and I didn't ask for reservedOnly, then just create one on-demand and return
	if len(reservationIDs) == 0 && !reservedOnly {
		device, _, err = svc.Create(createRequest)
		return device, err
	}

	// if we got here, we either had some reservation IDs, or we were asked to do reserved only.
	// In both cases, we try reservations first.
	for _, resID := range reservationIDs {
		createRequest.HardwareReservationID = resID
		device, _, err = svc.Create(createRequest)
		// if no error, we got the device, return it
		if err == nil {
			return device, err
		}
	}
	// if we got here, we failed to get a device with the given hardware reservation
	if reservedOnly {
		return nil, errors.New("could not get a device with the provided reservation IDs, and reservedOnly is true")
	}
	// now just create a device on demand
	device, _, err = svc.Create(createRequest)
	return device, err
}

func validateSecretAPIKey(secret *corev1.Secret) error {
	return validateSecret(secret, validation.SecretFieldAPIKey)
}
func validateSecret(secret *corev1.Secret, fields ...string) error {
	validationErr := validation.ValidateSecret(secret, fields...)
	if validationErr.ToAggregate() != nil && len(validationErr.ToAggregate().Errors()) > 0 {
		err := fmt.Errorf("Error while validating Secret %v", validationErr.ToAggregate().Error())
		klog.V(2).Infof("Validation of Secret failed %s", err)

		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func encodeMachineID(device *packngo.Device) string {
	return fmt.Sprintf("equinixmetal://%s/%s", device.Facility.Code, device.ID)
}

func decodeMachineID(id string) string {
	splitProviderID := strings.Split(id, "/")
	return splitProviderID[len(splitProviderID)-1]
}
