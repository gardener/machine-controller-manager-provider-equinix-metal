/*
Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved.
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
package provider_test

import (
	"fmt"
	"testing"

	"github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	testNamespace = "test"
)

func TestProvider(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provider Suite")
}

func newMachine(
	setMachineIndex int,
) *v1alpha1.Machine {
	index := 0

	if setMachineIndex > 0 {
		index = setMachineIndex
	}

	machine := &v1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "machine.sapcloud.io",
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("machine-%d", index),
			Namespace: testNamespace,
		},
	}

	// Don't initialize providerID and node if setMachineIndex == -1
	if setMachineIndex != -1 {
		machine.Spec = v1alpha1.MachineSpec{
			ProviderID: fmt.Sprintf("equinixmetal://ewr1/%06d", setMachineIndex),
		}

		machine.Labels=make(map[string]string)
		machine.Labels["node"]=fmt.Sprintf("machine-%d",setMachineIndex)
	}

	return machine
}

func newMachineClass(providerSpec []byte) *v1alpha1.MachineClass {
	return &v1alpha1.MachineClass{
		ProviderSpec: runtime.RawExtension{
			Raw: providerSpec,
		},
		Provider: provider.ProviderEquinixMetal,
	}
}

func setProvider(machine *v1alpha1.MachineClass, provider string) *v1alpha1.MachineClass {
	machine.Provider = provider
	return machine
}
