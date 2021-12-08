package driver

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/thingio/edge-device-std/config"
	"github.com/thingio/edge-device-std/logger"
	"github.com/thingio/edge-device-std/models"
	bus "github.com/thingio/edge-device-std/msgbus"
	"github.com/thingio/edge-device-std/operations"
	"sync"
)

func NewDeviceDriver(ctx context.Context, cancel context.CancelFunc,
	protocol *models.Protocol, twinBuilder models.DeviceTwinBuilder) (*DeviceDriver, error) {
	if protocol == nil {
		return nil, fmt.Errorf("the product cannot be nil")
	}
	if twinBuilder == nil {
		return nil, fmt.Errorf("please implement and specify the connector builder")
	}
	dd := &DeviceDriver{
		protocol:    protocol,
		twinBuilder: twinBuilder,

		products:    sync.Map{},
		devices:     sync.Map{},
		deviceTwins: sync.Map{},

		ctx:    ctx,
		cancel: cancel,
		logger: logger.NewLogger(),
	}

	return dd, nil
}

type DeviceDriver struct {
	// driver
	protocol    *models.Protocol
	twinBuilder models.DeviceTwinBuilder

	// caches
	products    sync.Map
	devices     sync.Map
	deviceTwins sync.Map

	// operation clients
	propsBus chan *models.DeviceDataWrapper
	eventBus chan *models.DeviceDataWrapper
	dc       operations.DriverClient
	ds       operations.DriverService

	// lifetime control variables for the device driver
	ctx    context.Context
	cancel context.CancelFunc
	logger *logger.Logger
	cfg    *config.Configuration
}

func (d *DeviceDriver) Initialize() error {
	if cfg, err := config.NewConfiguration(); err != nil {
		return err
	} else {
		d.cfg = cfg
	}

	if err := d.initializeOperations(); err != nil {
		return err
	}
	return nil
}

func (d *DeviceDriver) initializeOperations() error {
	d.propsBus = make(chan *models.DeviceDataWrapper, 1000)
	d.eventBus = make(chan *models.DeviceDataWrapper, 1000)

	mb, err := bus.NewMessageBus(&d.cfg.MessageBus, d.logger)
	if err != nil {
		return errors.Wrap(err, "fail to initialize the message bus")
	}

	dc, err := operations.NewDriverClient(mb, d.logger)
	if err != nil {
		return errors.Wrap(err, "fail to new an operations client")
	}
	d.dc = dc
	ds, err := operations.NewDriverService(mb, d.logger)
	if err != nil {
		return errors.Wrap(err, "fail to new an operations service")
	}
	d.ds = ds

	return nil
}

func (d *DeviceDriver) Serve() error {
	if err := d.subscribeMetaMutation(); err != nil {
		panic(err)
	}

	d.activateDevices()
	defer d.deactivateDevices()

	if err := d.handleDataOperation(); err != nil {
		panic(err)
	}
	go d.reportingDriverHealth()
	go d.reportingDevicesHealth()
	go d.reportingDevicesData()

	<-d.ctx.Done()
	return nil
}

func (d *DeviceDriver) putProduct(product *models.Product) {
	d.products.Store(product.ID, product)
}

func (d *DeviceDriver) getProduct(productID string) (*models.Product, error) {
	v, ok := d.products.Load(productID)
	if ok {
		return v.(*models.Product), nil
	}
	return nil, fmt.Errorf("the product[%s] is not found in cache", productID)
}

func (d *DeviceDriver) deleteProduct(productID string) {
	d.products.Delete(productID)
}

func (d *DeviceDriver) putDevice(device *models.Device) {
	d.devices.Store(device.ID, device)
}

func (d *DeviceDriver) getDevice(deviceID string) (*models.Device, error) {
	v, ok := d.devices.Load(deviceID)
	if ok {
		return v.(*models.Device), nil
	}
	return nil, fmt.Errorf("the device[%s] is not found in cache", deviceID)
}

func (d *DeviceDriver) deleteDevice(deviceID string) {
	d.devices.Delete(deviceID)
}

func (d *DeviceDriver) putDeviceTwin(deviceID string, deviceTwin models.DeviceTwin) {
	d.deviceTwins.Store(deviceID, deviceTwin)
}

func (d *DeviceDriver) getDeviceTwin(deviceID string) (models.DeviceTwin, error) {
	v, ok := d.deviceTwins.Load(deviceID)
	if ok {
		return v.(models.DeviceTwin), nil
	}
	return nil, fmt.Errorf("the device[%s] is not activated", deviceID)
}

func (d *DeviceDriver) deleteDeviceTwin(deviceID string) {
	d.deviceTwins.Delete(deviceID)
}