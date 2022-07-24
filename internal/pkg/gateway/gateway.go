package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/essentialkaos/translit/v2"
	"github.com/hotid/healthbox-hass-gw/pkg/healthbox"
	"github.com/hotid/healthbox-hass-gw/pkg/homeassistant"
	"sort"
	"strings"
	"time"
)

type Gw struct {
	devices map[string]*HaDevice
	state   bool
	ctx     context.Context
	cancel  context.CancelFunc
}

var ha *homeassistant.Client
var c *healthbox.Client

type HaDevice struct {
	HealthboxRoomId int
	Name            string
	SensorUniqueId  string
	SwitchUniqueId  string
	Identifiers     string
	State           string
	Unit            string
	Boost           healthbox.BoostInfo
	lastDiscovery   time.Time
	lastValue       time.Time
	lastBoostUpdate time.Time
}

func (g *Gw) StartGateway(ctx context.Context, healthboxClient *healthbox.Client, mqttClient *homeassistant.Client) {
	g.devices = make(map[string]*HaDevice)

	ha = mqttClient
	c = healthboxClient

	g.StartHaRestartHandler(mqttClient)
	g.Start()

	for {
		select {
		case <-ctx.Done():
			g.cancel()
			return
		}
	}
}

func (g *Gw) Start() {
	fmt.Println("starting gateway")
	g.ctx, g.cancel = context.WithCancel(context.Background())
	g.StartDataUpdate(g.ctx)
	g.resetDiscoveryInterval()
	g.StartDiscoveryPublishing(g.ctx)
}

func (g *Gw) Stop() {
	fmt.Println("stopping gateway")
	g.cancel()
}

func (d *HaDevice) SetState(newState string) {
	d.State = newState
	d.lastValue = time.Now()
	d.PublishAvailability()
	d.PublishState()
}

func (d *HaDevice) PublishState() {
	d.PublishSensorState()
	d.PublishSwitchState()
}

func (d *HaDevice) PublishSensorState() {
	ha.Publish(d.GetSensorStateTopic(), d.State)
}

func (d *HaDevice) PublishSwitchState() {
	ha.Publish(d.GetSwitchStateTopic(), d.getSwitchState())
}

func (d *HaDevice) PublishAvailability() {
	available := "0"
	if time.Since(d.lastValue) < 60*time.Second {
		available = "1"
	}
	ha.Publish(d.GetSensorAvailabilityTopic(), available)
	ha.Publish(d.GetSwitchAvailabilityTopic(), available)
}

func (d *HaDevice) PublishFlowSensorDiscovery() {
	discoveryInfo := homeassistant.HaSensorDiscoveryInfo{
		HaCommonDiscoveryInfo: homeassistant.HaCommonDiscoveryInfo{
			UniqueId:            d.SensorUniqueId,
			ObjectId:            d.SensorUniqueId,
			Name:                d.Name,
			PayloadAvailable:    "1",
			PayloadNotAvailable: "0",
			AvailabilityTopic:   d.GetSensorAvailabilityTopic(),
			StateTopic:          d.GetSensorStateTopic(),
		},
		UnitOfMeasurement: d.Unit,
		StateClass:        "measurement",
	}
	discoveryInfo.Device.Name = "Healthbox"
	discoveryInfo.Device.Identifiers = "healthbox"
	payload, err := json.Marshal(discoveryInfo)
	if err != nil {
		fmt.Printf("error marshalling discoveryInfo: %s", err)
		return
	}

	//fmt.Printf("publishing discovery data for device %s: %s %+v", d.Name, topic, string(payload))
	ha.Publish(fmt.Sprintf("homeassistant/sensor/%s/%s/config", "healthbox", d.SensorUniqueId), payload)
	d.lastDiscovery = time.Now()
}

func (d *HaDevice) PublishBoostSwitchDiscovery() {
	discoveryInfo := homeassistant.HaSwitchDiscoveryInfo{
		HaCommonDiscoveryInfo: homeassistant.HaCommonDiscoveryInfo{
			UniqueId:            d.SwitchUniqueId,
			ObjectId:            d.SwitchUniqueId,
			Name:                d.Name,
			PayloadAvailable:    "1",
			PayloadNotAvailable: "0",
			AvailabilityTopic:   d.GetSwitchAvailabilityTopic(),
			StateTopic:          d.GetSwitchStateTopic(),
		},
		CommandTopic: d.GetCommandTopic(),
		PayloadOn:    "1",
		PayloadOff:   "0",
		StateOn:      "1",
		StateOff:     "0",
	}
	discoveryInfo.Device.Name = "Healthbox"
	discoveryInfo.Device.Identifiers = "healthbox"

	payload, err := json.Marshal(discoveryInfo)
	if err != nil {
		fmt.Printf("error marshalling discoveryInfo: %s", err)
		return
	}

	ha.Publish(fmt.Sprintf("homeassistant/switch/%s/%s/config", "healthbox", d.SwitchUniqueId), payload)
	d.lastDiscovery = time.Now()
}

func (g *Gw) String() string {
	var devicesStr string
	keys := make([]string, 0, len(g.devices))
	for k := range g.devices {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return g.devices[keys[i]].Name < g.devices[keys[j]].Name
	})
	for _, key := range keys {
		device := g.devices[key]
		devicesStr += fmt.Sprintf("%s: %s %s (boost: %t)\n", device.Name, device.State, device.Unit, device.Boost.Enable)
	}
	return devicesStr
}

func (g *Gw) UpdateCurrentData(currentData *healthbox.CurrentData) {
	for _, room := range currentData.Room {
		state := fmt.Sprintf("%.1f", room.Actuator[0].Parameter.FlowRate.Value)
		deviceId := getUniqueSensorId(room)

		if device, ok := g.devices[deviceId]; !ok {
			g.devices[deviceId] = newHaDevice(room)
		} else {
			device.SetState(state)
		}
	}
}

func (g *Gw) UpdateRoomsBoostStatus() {
	for _, room := range g.devices {
		if time.Since(room.lastBoostUpdate) < 5*time.Second {
			continue
		}

		boostInfo, err := c.GetBoostInfo(room.HealthboxRoomId)
		if err != nil {
			continue
		}

		room.Boost = *boostInfo
		room.lastBoostUpdate = time.Now()
	}
}

func newHaDevice(room healthbox.RoomInfo) *HaDevice {
	state := fmt.Sprintf("%.1f", room.Actuator[0].Parameter.FlowRate.Value)

	device := HaDevice{
		HealthboxRoomId: room.Id,
		Name:            room.Name,
		SensorUniqueId:  getUniqueSensorId(room),
		SwitchUniqueId:  getUniqueSwitchId(room),
		Identifiers:     "healthbox",
		State:           state,
		Unit:            room.Actuator[0].Parameter.FlowRate.Unit,
		lastValue:       time.Now(),
		lastDiscovery:   time.Now(),
	}
	device.PublishFlowSensorDiscovery()
	device.PublishBoostSwitchDiscovery()
	ha.Mqtt.Subscribe(device.GetCommandTopic(), 0, device.mqttSwitchCallback)
	return &device
}

func transliterateName(name string) string {
	name = strings.Replace(translit.EncodeToICAO(name), " ", "_", -1)
	name = strings.Replace(name, "(", "", -1)
	name = strings.Replace(name, ")", "", -1)
	return name
}

func getUniqueSensorId(room healthbox.RoomInfo) string {
	return fmt.Sprintf("%s_%s_%s", "healthbox", transliterateName(room.Name), "air_flow")
}

func getUniqueSwitchId(room healthbox.RoomInfo) string {
	return fmt.Sprintf("%s_%s_%s", "healthbox", transliterateName(room.Name), "boost")
}

func (g *Gw) StartDataUpdate(ctx context.Context) {
	go func() {
		for {
			currentData, err := c.GetCurrentData()
			if err != nil {
				fmt.Printf("error getting current data: %s", err)
				goto NextCycle
			}

			g.UpdateCurrentData(currentData)
			g.UpdateRoomsBoostStatus()
			fmt.Printf("%s", g)

		NextCycle:
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 15)
			}
		}
	}()
}
func (g *Gw) resetDiscoveryInterval() {
	for _, device := range g.devices {
		device.lastDiscovery = time.Time{}
	}
}

func (g *Gw) StartDiscoveryPublishing(ctx context.Context) {
	go func() {
		for {
			for _, device := range g.devices {
				if time.Since(device.lastDiscovery) > 3600*time.Second {
					device.PublishFlowSensorDiscovery()
				}
			}
			select {
			case <-ctx.Done():
				fmt.Printf("discovery publishing task exiting")
				return
			default:
				time.Sleep(time.Second * 15)
			}
		}
	}()
}

func (g *Gw) StartHaRestartHandler(client *homeassistant.Client) {
	handler := g.haRestartHandler(client)
	client.Mqtt.Subscribe(client.StatusTopic(), 0, handler)
}

func (g *Gw) haRestartHandler(haClient *homeassistant.Client) mqtt.MessageHandler {
	return func(client mqtt.Client, message mqtt.Message) {
		if fmt.Sprintf("%s", message.Payload()) == haClient.PayloadOffline() {
			g.Stop()
		} else {
			g.Start()
		}
	}
}

func (d *HaDevice) GetSensorAvailabilityTopic() string {
	return fmt.Sprintf("%s/devices/controls/%s/availability", "healthbox", d.SensorUniqueId)
}

func (d *HaDevice) GetSensorStateTopic() string {
	return fmt.Sprintf("%s/devices/controls/%s", "healthbox", d.SensorUniqueId)
}

func (d *HaDevice) GetSwitchAvailabilityTopic() string {
	return fmt.Sprintf("%s/devices/controls/%s/availability", "healthbox", d.SwitchUniqueId)
}

func (d *HaDevice) GetSwitchStateTopic() string {
	return fmt.Sprintf("%s/devices/controls/%s", "healthbox", d.SwitchUniqueId)
}

func (d *HaDevice) getSwitchState() string {
	if d.Boost.Enable {
		return "1"
	}
	return "0"
}

func (d *HaDevice) GetCommandTopic() string {
	return fmt.Sprintf("%s/%s", d.GetSwitchStateTopic(), "set")
}

func (d *HaDevice) mqttSwitchCallback(client mqtt.Client, message mqtt.Message) {

	if fmt.Sprintf("%s", message.Payload()) == "0" {
		d.SetBoost(false)
	} else {
		d.SetBoost(true)
	}

}

func (d *HaDevice) SetBoost(newBoostState bool) {
	if d.Boost.Enable == newBoostState {
		return
	}

	boostCommand := healthbox.BoostInfo{Enable: newBoostState, Level: 200, Timeout: 3600}

	payload, err := json.Marshal(boostCommand)
	if err != nil {
		panic(err)
	}

	err = c.Put(d.getBoostUrl(), payload)
	if err != nil {
		fmt.Printf("error setting boost level: %s", err)
	}

	time.Sleep(100 * time.Millisecond)

	boostInfo, err := c.GetBoostInfo(d.HealthboxRoomId)
	if err != nil {
		return
	}

	d.Boost = *boostInfo
	d.lastBoostUpdate = time.Now()

	d.PublishState()
}

func (d *HaDevice) getBoostUrl() string {
	return fmt.Sprintf("/v1/api/boost/%d", d.HealthboxRoomId)
}
