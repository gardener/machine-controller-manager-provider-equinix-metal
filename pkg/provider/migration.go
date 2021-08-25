package provider

import (
	"encoding/json"

	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"k8s.io/apimachinery/pkg/runtime"
)

// fillUpMachineClass copies over the fields from ProviderMachineClass to MachineClass
func fillUpMachineClass(packetMachineClass *v1alpha1.PacketMachineClass, machineClass *v1alpha1.MachineClass) error {
	// Prepare the providerSpec struct
	providerSpec := &api.EquinixMetalProviderSpec{
		APIVersion:   api.V1alpha1,
		MachineType:  packetMachineClass.Spec.MachineType,
		BillingCycle: packetMachineClass.Spec.BillingCycle,
		OS:           packetMachineClass.Spec.OS,
		ProjectID:    packetMachineClass.Spec.ProjectID,
		SSHKeys:      packetMachineClass.Spec.SSHKeys,
		UserData:     packetMachineClass.Spec.UserData,
	}

	// make sure to copy slices
	providerSpec.Facilities = make([]string, len(packetMachineClass.Spec.Facility))
	copy(providerSpec.Facilities, packetMachineClass.Spec.Facility)
	providerSpec.Tags = make([]string, len(packetMachineClass.Spec.Tags))
	copy(providerSpec.Tags, packetMachineClass.Spec.Tags)
	providerSpec.SSHKeys = make([]string, len(packetMachineClass.Spec.SSHKeys))
	copy(providerSpec.SSHKeys, packetMachineClass.Spec.SSHKeys)

	// Marshal providerSpec into Raw Bytes
	providerSpecMarshal, err := json.Marshal(providerSpec)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	// Migrate finalizers, labels, annotations
	machineClass.Name = packetMachineClass.Name
	machineClass.Labels = packetMachineClass.Labels
	machineClass.Annotations = packetMachineClass.Annotations
	machineClass.Finalizers = packetMachineClass.Finalizers
	machineClass.ProviderSpec = runtime.RawExtension{
		Raw: providerSpecMarshal,
	}
	machineClass.SecretRef = packetMachineClass.Spec.SecretRef
	machineClass.CredentialsSecretRef = packetMachineClass.Spec.CredentialsSecretRef
	machineClass.Provider = ProviderEquinixMetal

	return nil
}
