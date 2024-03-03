package spi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/equinix/equinix-sdk-go/services/metalv1"
	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct {
}

// NewSession creates a session for equinix metal provider
func (p *PluginSPIImpl) NewSession(secret *corev1.Secret) (MetalDeviceService, error) {
	apiKey := GetAPIKey(secret)
	token := strings.TrimSpace(apiKey)

	if token == "" {
		return nil, errors.New("Equinix Metal api token required")
	}

	configuration := metalv1.NewConfiguration()
	configuration.Debug = true
	configuration.AddDefaultHeader("X-Auth-Token", token)
	client := metalv1.NewAPIClient(configuration)

	return &metalDeviceSvc{client: client}, nil
}

// GetAPIKey extracts the APIKey from the *corev1.Secret object
func GetAPIKey(secret *corev1.Secret) string {
	return extractCredentialsFromData(secret.Data, api.APIKey)
}

// extractCredentialsFromData extracts and trims a value from the given data map. The first key that exists is being
// returned, otherwise, the next key is tried, etc. If no key exists then an empty string is returned.
func extractCredentialsFromData(data map[string][]byte, keys ...string) string {
	for _, key := range keys {
		if val, ok := data[key]; ok {
			return strings.TrimSpace(string(val))
		}
	}
	return ""
}

type metalDeviceSvc struct {
	client *metalv1.APIClient
}

func (a *metalDeviceSvc) FindProjectDevices(
	ctx context.Context,
	projectID string,
) (*metalv1.DeviceList, *http.Response, error) {
	return a.client.DevicesApi.FindProjectDevices(ctx, projectID).Execute()
}

func (a *metalDeviceSvc) FindDeviceByID(
	ctx context.Context,
	deviceID string,
) (*metalv1.Device, *http.Response, error) {
	return a.client.DevicesApi.FindDeviceById(ctx, deviceID).Execute()
}

func (a *metalDeviceSvc) CreateDevice(
	ctx context.Context,
	projectID string,
	createDeviceRequest metalv1.CreateDeviceRequest,
) (*metalv1.Device, *http.Response, error) {
	return a.client.DevicesApi.
		CreateDevice(ctx, projectID).
		CreateDeviceRequest(createDeviceRequest).Execute()
}

func (a *metalDeviceSvc) DeleteDevice(
	ctx context.Context,
	deviceID string,
) (*http.Response, error) {
	return a.client.DevicesApi.DeleteDevice(ctx, deviceID).Execute()
}
