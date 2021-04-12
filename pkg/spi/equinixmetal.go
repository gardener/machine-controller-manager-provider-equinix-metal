package spi

import (
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	"github.com/packethost/packngo"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPIImpl is the real implementation of SPI interface that makes the calls to the provider SDK.
type PluginSPIImpl struct {
}

// NewSession creates a session for equinix metal provider
func (p *PluginSPIImpl) NewSession(secret *corev1.Secret) packngo.DeviceService {
	apiKey := GetAPIKey(secret)
	token := strings.TrimSpace(apiKey)

	if token != "" {
		client := packngo.NewClientWithAuth("gardener", token, nil)
		if client == nil {
			return nil
		}
		return client.Devices
	}
	return nil
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
