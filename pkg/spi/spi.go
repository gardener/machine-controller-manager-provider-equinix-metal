// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0


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
