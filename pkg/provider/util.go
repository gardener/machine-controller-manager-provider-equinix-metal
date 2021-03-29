package provider

import (
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/provider/apis"
	corev1 "k8s.io/api/core/v1"
)

func getApiKey(secret *corev1.Secret) string {
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
