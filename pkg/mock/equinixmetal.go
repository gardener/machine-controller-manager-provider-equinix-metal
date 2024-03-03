package mock

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/equinix/equinix-sdk-go/services/metalv1"
	"github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/spi"
	corev1 "k8s.io/api/core/v1"
)

// PluginSPIImpl is the plugin SPI implementation to mock the provider
type PluginSPIImpl struct {
	Devices []metalv1.Device
	index   int
	mu      sync.Mutex // so that we can increment index without conflicts
}

// NewSession creates a mock session for provider
func (p *PluginSPIImpl) NewSession(secret *corev1.Secret) (spi.MetalDeviceService, error) {
	apiKey := spi.GetAPIKey(secret)
	token := strings.TrimSpace(apiKey)

	if token == "" {
		return nil, errors.New("Equinix Metal api token required")
	}

	return &deviceService{
		spi:   p,
		name:  "gardener",
		token: token,
	}, nil
}

func (p *PluginSPIImpl) increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.index++
}

func (p *PluginSPIImpl) addDevice(dev metalv1.Device) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Devices = append(p.Devices, dev)
}

type deviceService struct {
	spi   *PluginSPIImpl
	name  string
	token string
}

func (d *deviceService) FindProjectDevices(
	ctx context.Context,
	projectID string,
) (*metalv1.DeviceList, *http.Response, error) {
	return &metalv1.DeviceList{
		Devices: d.spi.Devices,
	}, &http.Response{}, nil
}

func (d *deviceService) FindDeviceByID(
	ctx context.Context,
	deviceID string,
) (*metalv1.Device, *http.Response, error) {
	for _, dev := range d.spi.Devices {
		if *dev.Id == deviceID {
			return &dev, &http.Response{}, nil
		}
	}
	return nil, &http.Response{
		StatusCode: 404,
		Status:     "404 NOT FOUND",
	}, fmt.Errorf("404 NOT FOUND")
}

func (d *deviceService) CreateDevice(
	ctx context.Context,
	projectID string,
	createDeviceRequest metalv1.CreateDeviceRequest,
) (*metalv1.Device, *http.Response, error) {
	now := time.Now()
	d.spi.increment()
	req := createDeviceRequest.DeviceCreateInMetroInput
	var (
		name         = fmt.Sprintf("%06d", d.spi.index)
		billingCycle = string(*req.BillingCycle)
	)
	dev := metalv1.Device{
		Id:           &name,
		Hostname:     req.Hostname,
		Description:  req.Description,
		CreatedAt:    &now,
		UpdatedAt:    &now,
		BillingCycle: &billingCycle,
		Tags:         req.Tags,
		OperatingSystem: &metalv1.OperatingSystem{
			Name: &req.OperatingSystem,
		},
		Plan: &metalv1.Plan{},
		Metro: &metalv1.DeviceMetro{
			Code: &req.Metro,
		},
		Project: &metalv1.Project{
			Id: &projectID,
		},
		Userdata: req.Userdata,
	}
	d.spi.addDevice(dev)
	return &dev, &http.Response{}, nil
}

func (d *deviceService) DeleteDevice(
	ctx context.Context,
	deviceID string,
) (*http.Response, error) {
	var devs []metalv1.Device
	for _, dev := range d.spi.Devices {
		if *dev.Id != deviceID {
			devs = append(devs, dev)
		}
	}
	d.spi.Devices = devs
	return &http.Response{}, nil
}
