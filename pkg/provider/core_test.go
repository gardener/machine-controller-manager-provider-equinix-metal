package provider_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/mock"
	"github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider"
	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	messageApiKeyMissing        = "machine codes error: code = [InvalidArgument] message = [machine codes error: code = [Internal] message = [Error while validating Secret secretRef.apiToken: Required value: Required Equinix Metal API Key one of 'apiToken' or 'alternateApiToken']]"
	messageUserdataMissing      = "machine codes error: code = [InvalidArgument] message = [machine codes error: code = [Internal] message = [Error while validating Secret secretRef.userData: Required value: Required userData]]"
	messageWrongProvider        = "machine codes error: code = [InvalidArgument] message = [Requested for Provider '%s', we only support 'EquinixMetal']"
	messageNotFound             = "machine codes error: code = [NotFound] message = [Could not get device %s: 404 NOT FOUND]"
	messageVolumesUnimplemented = "machine codes error: code = [Unimplemented] message = [Equinix Metal does not have storage]"
)

var _ = Describe("MachineServer", func() {
	// Some initializations
	providerSpecStruct := api.EquinixMetalProviderSpec{
		Facilities:   []string{"ewr1", "ny5"},
		Metro:        "ny",
		MachineType:  "c3.small.x86",
		BillingCycle: "hourly",
		OS:           "alpine_3.13",
		ProjectID:    "abcdefg",
		Tags: []string{
			"kubernetes.io/cluster/shoot-test: 1",
			"kubernetes.io/role/test: 1",
		},
	}
	providerSpec, _ := json.Marshal(providerSpecStruct)
	providerSecret := &corev1.Secret{
		Data: map[string][]byte{
			"apiToken": []byte("dummy-token"),
			"userData": []byte("dummy-user-data"),
		},
	}

	Describe("#CreateMachine", func() {
		type setup struct {
		}
		type action struct {
			machineRequest *driver.CreateMachineRequest
		}
		type expect struct {
			machineResponse   *driver.CreateMachineResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()
				response, err := p.CreateMachine(ctx, data.action.machineRequest)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(data.expect.machineResponse.ProviderID).To(Equal(response.ProviderID))
					Expect(data.expect.machineResponse.NodeName).To(Equal(response.NodeName))
				}
			},
			Entry("simple", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					machineResponse: &driver.CreateMachineResponse{
						ProviderID: "equinixmetal://ewr1/000001",
						NodeName:   "machine-0",
					},
					errToHaveOccurred: false,
				},
			}),
			Entry("wrong provider", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: setProvider(newMachineClass(providerSpec), "badprovider"),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        fmt.Sprintf(messageWrongProvider, "badprovider"),
				},
			}),
			Entry("missing key", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{
								"userData": providerSecret.Data["userData"],
							},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageApiKeyMissing,
				},
			}),
			Entry("missing userData", &data{
				action: action{
					machineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{
								"apiToken": providerSecret.Data["apiToken"],
							},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageUserdataMissing,
				},
			}),
		)
	})
	Describe("#DeleteMachine", func() {
		type setup struct {
			createMachineRequest *driver.CreateMachineRequest
			resetProviderToEmpty bool
		}
		type action struct {
			deleteMachineRequest *driver.DeleteMachineRequest
		}
		type expect struct {
			deleteMachineResponse *driver.DeleteMachineResponse
			errToHaveOccurred     bool
			errMessage            string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()
				if data.setup.createMachineRequest != nil {
					_, err := p.CreateMachine(ctx, data.setup.createMachineRequest)
					Expect(err).ToNot(HaveOccurred())
				}
				_, err := p.DeleteMachine(ctx, data.action.deleteMachineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("existing machine", &data{
				setup: setup{
					createMachineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(0),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				action: action{
					deleteMachineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine(0),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					deleteMachineResponse: &driver.DeleteMachineResponse{},
					errToHaveOccurred:     false,
				},
			}),
			Entry("non-existing machine", &data{
				action: action{
					deleteMachineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine(999),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					deleteMachineResponse: &driver.DeleteMachineResponse{},
					errToHaveOccurred:     false,
				},
			}),
			Entry("wrong provider", &data{
				action: action{
					deleteMachineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine(999),
						MachineClass: setProvider(newMachineClass(providerSpec), "badprovider"),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        fmt.Sprintf(messageWrongProvider, "badprovider"),
				},
			}),
			Entry("missing key", &data{
				action: action{
					deleteMachineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageApiKeyMissing,
				},
			}),
			Entry("missing userData", &data{
				action: action{
					deleteMachineRequest: &driver.DeleteMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{
								"apiToken": providerSecret.Data["apiToken"],
							},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: false,
				},
			}),
		)
	})
	Describe("#GetMachineStatus", func() {
		type setup struct {
			createMachineRequest *driver.CreateMachineRequest
		}
		type action struct {
			getMachineRequest *driver.GetMachineStatusRequest
		}
		type expect struct {
			getMachineResponse *driver.GetMachineStatusResponse
			errToHaveOccurred  bool
			errMessage         string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()
				if data.setup.createMachineRequest != nil {
					_, err := p.CreateMachine(ctx, data.setup.createMachineRequest)
					Expect(err).ToNot(HaveOccurred())
				}
				_, err := p.GetMachineStatus(ctx, data.action.getMachineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("existing machine", &data{
				setup: setup{
					createMachineRequest: &driver.CreateMachineRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				action: action{
					getMachineRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine(1),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{},
			}),
			Entry("non-existing machine", &data{
				action: action{
					getMachineRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine(0),
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        fmt.Sprintf(messageNotFound, "000000"),
				},
			}),
			Entry("wrong provider", &data{
				action: action{
					getMachineRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine(999),
						MachineClass: setProvider(newMachineClass(providerSpec), "badprovider"),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        fmt.Sprintf(messageWrongProvider, "badprovider"),
				},
			}),
			Entry("missing key", &data{
				action: action{
					getMachineRequest: &driver.GetMachineStatusRequest{
						Machine:      newMachine(-1),
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageApiKeyMissing,
				},
			}),
		)
	})
	Describe("#ListMachines", func() {
		type setup struct {
			createMachineRequest []*driver.CreateMachineRequest
		}
		type action struct {
			listMachineRequest *driver.ListMachinesRequest
		}
		type expect struct {
			listMachineResponse *driver.ListMachinesResponse
			errToHaveOccurred   bool
			errMessage          string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()
				for _, createReq := range data.setup.createMachineRequest {
					_, err := p.CreateMachine(ctx, createReq)
					Expect(err).ToNot(HaveOccurred())
				}
				listResponse, err := p.ListMachines(ctx, data.action.listMachineRequest)

				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(listResponse.MachineList)).To(Equal(len(data.expect.listMachineResponse.MachineList)))
				}
			},
			Entry("simple", &data{
				setup: setup{
					createMachineRequest: []*driver.CreateMachineRequest{
						{
							Machine:      newMachine(0),
							MachineClass: newMachineClass(providerSpec),
							Secret:       providerSecret,
						},
						{
							Machine:      newMachine(1),
							MachineClass: newMachineClass(providerSpec),
							Secret:       providerSecret,
						},
						{
							Machine:      newMachine(2),
							MachineClass: newMachineClass(providerSpec),
							Secret:       providerSecret,
						},
					},
				},
				action: action{
					listMachineRequest: &driver.ListMachinesRequest{
						MachineClass: newMachineClass(providerSpec),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: false,
					listMachineResponse: &driver.ListMachinesResponse{
						MachineList: map[string]string{
							"equinixmetal:///ewr1/000000": "machine-0",
							"equinixmetal:///ewr1/000001": "machine-1",
							"equinixmetal:///ewr1/000002": "machine-2",
						},
					},
				},
			}),
			Entry("wrong provider", &data{
				action: action{
					listMachineRequest: &driver.ListMachinesRequest{
						MachineClass: setProvider(newMachineClass(providerSpec), "badprovider"),
						Secret:       providerSecret,
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        fmt.Sprintf(messageWrongProvider, "badprovider"),
				},
			}),
			Entry("missing key", &data{
				action: action{
					listMachineRequest: &driver.ListMachinesRequest{
						MachineClass: newMachineClass(providerSpec),
						Secret: &corev1.Secret{
							Data: map[string][]byte{},
						},
					},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageApiKeyMissing,
				},
			}),
		)
	})
	Describe("#GetVolumeIDs", func() {
		type setup struct {
		}
		type action struct {
			request *driver.GetVolumeIDsRequest
		}
		type expect struct {
			response          *driver.GetVolumeIDsResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()
				_, err := p.GetVolumeIDs(ctx, data.action.request)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
				}
			},
			Entry("simple", &data{
				action: action{
					request: &driver.GetVolumeIDsRequest{},
				},
				expect: expect{
					errToHaveOccurred: true,
					errMessage:        messageVolumesUnimplemented,
				},
			}),
		)
	})
	Describe("#GenerateMachineClassForMigration", func() {
		type setup struct {
		}
		type action struct {
			generateMachineClassForMigrationRequest *driver.GenerateMachineClassForMigrationRequest
		}
		type expect struct {
			machineClass *v1alpha1.MachineClass
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}
		/*
			SSHKeys      []string `json:"sshKeys,omitempty"`
			UserData     string   `json:"userdata,omitempty"`
		*/
		validRaw := `{"apiVersion":"mcm.gardener.cloud/v1alpha1","facility":["ewr1"],"machineType":"c3.medium.x86","billingCycle":"hourly","OS":"ubuntu_2004","projectID":"abcdefg","tags":["key1: value1","key2: value2"],"userdata":"dummy-user-data"}`
		DescribeTable("##table",
			func(data *data) {
				plugin := &mock.PluginSPIImpl{}
				p := provider.NewProvider(plugin)
				ctx := context.Background()

				_, _ = p.GenerateMachineClassForMigration(
					ctx,
					data.action.generateMachineClassForMigrationRequest,
				)

				fmt.Println(string(data.action.generateMachineClassForMigrationRequest.MachineClass.ProviderSpec.Raw))
				fmt.Println(string(data.expect.machineClass.ProviderSpec.Raw))
				Expect(data.action.generateMachineClassForMigrationRequest.MachineClass).To(Equal(data.expect.machineClass))
			},
			Entry("Simple migration request with all fields set", &data{
				action: action{
					generateMachineClassForMigrationRequest: &driver.GenerateMachineClassForMigrationRequest{
						ProviderSpecificMachineClass: &v1alpha1.PacketMachineClass{
							ObjectMeta: v1.ObjectMeta{
								Name: "test-mc",
								Labels: map[string]string{
									"key1": "value1",
									"key2": "value2",
								},
								Annotations: map[string]string{
									"key1": "value1",
									"key2": "value2",
								},
								Finalizers: []string{
									"mcm/finalizer",
								},
							},
							TypeMeta: v1.TypeMeta{},
							Spec: v1alpha1.PacketMachineClassSpec{
								Facility:     []string{"ewr1"},
								MachineType:  "c3.medium.x86",
								BillingCycle: "hourly",
								OS:           "ubuntu_2004",
								ProjectID:    "abcdefg",
								UserData:     "dummy-user-data",
								Tags: []string{
									"key1: value1",
									"key2: value2",
								},
								SecretRef: &corev1.SecretReference{
									Name:      "test-secret",
									Namespace: "test-namespace",
								},
								CredentialsSecretRef: &corev1.SecretReference{
									Name:      "test-credentials",
									Namespace: "test-namespace",
								},
							},
						},
						MachineClass: &v1alpha1.MachineClass{},
						ClassSpec: &v1alpha1.ClassSpec{
							Kind: provider.PacketMachineClassKind,
							Name: "test-mc",
						},
					},
				},
				expect: expect{
					machineClass: &v1alpha1.MachineClass{
						TypeMeta: v1.TypeMeta{},
						ObjectMeta: v1.ObjectMeta{
							Name: "test-mc",
							Labels: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
							Finalizers: []string{
								"mcm/finalizer",
							},
						},
						ProviderSpec: runtime.RawExtension{
							Raw: []byte(validRaw),
						},
						SecretRef: &corev1.SecretReference{
							Name:      "test-secret",
							Namespace: "test-namespace",
						},
						CredentialsSecretRef: &corev1.SecretReference{
							Name:      "test-credentials",
							Namespace: "test-namespace",
						},
						Provider: provider.ProviderEquinixMetal,
					},
				},
			}),
		)
	})

})
