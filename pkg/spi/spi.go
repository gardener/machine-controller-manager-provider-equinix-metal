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

package spi

import (
	"context"
	"net/http"

	"github.com/equinix/equinix-sdk-go/services/metalv1"
	corev1 "k8s.io/api/core/v1"
)

// MetalDeviceService is a simple interface for the metalv1 device api.
// It only contains simplified api that is required for the machine controller.
type MetalDeviceService interface {
	FindProjectDevices(ctx context.Context, projectID string) (*metalv1.DeviceList, *http.Response, error)
	FindDeviceByID(ctx context.Context, deviceID string) (*metalv1.Device, *http.Response, error)
	CreateDevice(
		ctx context.Context,
		projectID string,
		createDeviceRequest metalv1.CreateDeviceRequest,
	) (*metalv1.Device, *http.Response, error)
	DeleteDevice(ctx context.Context, deviceID string) (*http.Response, error)
}

// SessionProviderInterface provides an interface to deal with cloud provider session
// Example interfaces are listed below.
type SessionProviderInterface interface {
	NewSession(*corev1.Secret) (MetalDeviceService, error)
}
