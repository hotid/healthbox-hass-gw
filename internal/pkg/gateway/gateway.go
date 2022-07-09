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

type GwDevices map[string]*HaDevice

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
}

func (d GwDevices) StartGateway(ctx context.Context, healthboxClient *healthbox.Client, mqttClient *homeassistant.Client) {
	ha = mqttClient
	c = healthboxClient
	d.StartDataUpdate(ctx, healthboxClient)
	d.StartDiscoveryPublishing(ctx)
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
	if time.Since(d.lastValue) < time.Duration(60*time.Second) {
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

func (d GwDevices) String() string {
	var devicesStr string
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return d[keys[i]].Name < d[keys[j]].Name
	})
	for _, key := range keys {
		device := d[key]
		devicesStr += fmt.Sprintf("%s: %s %s (boost: %t)\n", device.Name, device.State, device.Unit, device.Boost.Enable)
	}
	return devicesStr
}

func (d GwDevices) UpdateCurrentData(currentData *healthbox.CurrentData) {
	for _, room := range currentData.Room {
		state := fmt.Sprintf("%.1f", room.Actuator[0].Parameter.FlowRate.Value)
		deviceId := getUniqueSensorId(room)

		if device, ok := d[deviceId]; !ok {
			d[deviceId] = newHaDevice(room)
		} else {
			device.SetState(state)
		}
	}
}

func (d GwDevices) UpdateBoostStatus(boostInfo map[int]healthbox.BoostInfo) {
	for _, room := range d {
		if roomBoostInfo, ok := boostInfo[room.HealthboxRoomId]; ok {
			room.Boost = roomBoostInfo
		}
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

func (d GwDevices) StartDataUpdate(ctx context.Context, c *healthbox.Client) {
	go func() {
		for {
			var boostInfo *map[int]healthbox.BoostInfo
			currentData, err := c.GetCurrentData()
			if err != nil {
				fmt.Printf("error getting current data: %s", err)
				goto NextCycle
			}

			d.UpdateCurrentData(currentData)
			boostInfo, err = c.GetBoostInfo(currentData)
			if err != nil {
				fmt.Printf("error getting boost info: %s", err)
				goto NextCycle
			}

			d.UpdateBoostStatus(*boostInfo)
			fmt.Printf("%s", d)

		NextCycle:
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Second * 1)
			}
		}
	}()
}

func (d GwDevices) StartDiscoveryPublishing(ctx context.Context) {
	go func() {
		for {
			for _, device := range d {
				if time.Since(device.lastDiscovery) > time.Duration(3600*time.Second) {
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

func (d *HaDevice) SetBoost(enabled bool) {

	boostCommand := healthbox.BoostInfo{Enable: enabled, Level: 200, Timeout: 3600}

	payload, err := json.Marshal(boostCommand)
	if err != nil {
		panic(err)
	}

	err = c.Put(d.getBoostUrl(), payload)
	if err != nil {
		fmt.Printf("error setting boost level: %s", err)
	}
	d.Boost.Enable = enabled
	d.PublishState()
}

func (d *HaDevice) getBoostUrl() string {
	return fmt.Sprintf("/v1/api/boost/%d", d.HealthboxRoomId)
}
