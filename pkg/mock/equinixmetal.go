package mock

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gardener/machine-controller-manager-provider-equinix-metal/pkg/spi"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	"github.com/packethost/packngo"
	corev1 "k8s.io/api/core/v1"
)

type PluginSPIImpl struct {
	Devices []packngo.Device
	index   int
	mu      sync.Mutex // so that we can increment index without conflicts
}

func (p *PluginSPIImpl) NewSession(secret *corev1.Secret) packngo.DeviceService {
	apiKey := spi.GetApiKey(secret)
	token := strings.TrimSpace(apiKey)

	if token != "" {
		return &deviceService{
			spi:   p,
			name:  "gardener",
			token: token,
		}
	}
	return nil
}

func (p *PluginSPIImpl) increment() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.index++
}

func (p *PluginSPIImpl) addDevice(dev packngo.Device) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Devices = append(p.Devices, dev)
}

type deviceService struct {
	spi   *PluginSPIImpl
	name  string
	token string
}

func (d *deviceService) List(projectID string, opts *packngo.ListOptions) ([]packngo.Device, *packngo.Response, error) {
	return d.spi.Devices, &packngo.Response{}, nil
}
func (d *deviceService) Get(deviceID string, opts *packngo.GetOptions) (*packngo.Device, *packngo.Response, error) {
	for _, dev := range d.spi.Devices {
		if dev.ID == deviceID {
			return &dev, &packngo.Response{}, nil
		}
	}
	return nil, &packngo.Response{
		Response: &http.Response{
			StatusCode: 404,
			Status:     "404 NOT FOUND",
		},
	}, fmt.Errorf("404 NOT FOUND")
}
func (d *deviceService) Create(req *packngo.DeviceCreateRequest) (*packngo.Device, *packngo.Response, error) {
	now := time.Now()
	d.spi.increment()
	dev := packngo.Device{
		ID:           fmt.Sprintf("%06d", d.spi.index),
		Hostname:     req.Hostname,
		Description:  &req.Description,
		Created:      now.String(),
		Updated:      now.String(),
		BillingCycle: req.BillingCycle,
		Tags:         req.Tags,
		OS: &packngo.OS{
			Name: req.OS,
		},
		Plan: &packngo.Plan{},
		Facility: &packngo.Facility{
			Code: req.Facility[0],
		},
		Project: &packngo.Project{
			ID: req.ProjectID,
		},
		UserData: req.UserData,
	}
	d.spi.addDevice(dev)
	return &dev, &packngo.Response{}, nil
}
func (d *deviceService) Delete(deviceID string, force bool) (*packngo.Response, error) {
	var devs []packngo.Device
	for _, dev := range d.spi.Devices {
		if dev.ID != deviceID {
			devs = append(devs, dev)
		}
	}
	d.spi.Devices = devs
	return &packngo.Response{}, nil
}

/*
 Below are not implemenetd as unnecessary for MCM
*/
func (d *deviceService) Update(string, *packngo.DeviceUpdateRequest) (*packngo.Device, *packngo.Response, error) {
	return nil, nil, status.Error(codes.Unimplemented, "Update unsupported in mock")
}
func (d *deviceService) Reboot(string) (*packngo.Response, error) {
	return nil, status.Error(codes.Unimplemented, "Reboot unsupported in mock")
}
func (d *deviceService) PowerOff(string) (*packngo.Response, error) {
	return nil, status.Error(codes.Unimplemented, "PowerOff unsupported in mock")
}
func (d *deviceService) PowerOn(string) (*packngo.Response, error) {
	return nil, status.Error(codes.Unimplemented, "PowerOn unsupported in mock")
}
func (d *deviceService) Lock(string) (*packngo.Response, error) {
	return nil, status.Error(codes.Unimplemented, "Lock unsupported in mock")
}
func (d *deviceService) Unlock(string) (*packngo.Response, error) {
	return nil, status.Error(codes.Unimplemented, "Unlock unsupported in mock")
}
func (d *deviceService) ListBGPSessions(deviceID string, opts *packngo.ListOptions) ([]packngo.BGPSession, *packngo.Response, error) {
	return nil, nil, status.Error(codes.Unimplemented, "ListBGPSessions unsupported in mock")
}
func (d *deviceService) ListBGPNeighbors(deviceID string, opts *packngo.ListOptions) ([]packngo.BGPNeighbor, *packngo.Response, error) {
	return nil, nil, status.Error(codes.Unimplemented, "ListBGPNeighbors unsupported in mock")
}
func (d *deviceService) ListEvents(deviceID string, opts *packngo.ListOptions) ([]packngo.Event, *packngo.Response, error) {
	return nil, nil, status.Error(codes.Unimplemented, "ListEvents unsupported in mock")
}
