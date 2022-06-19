package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/essentialkaos/translit/v2"
	"github.com/hotid/healthbox-hass-gw/pkg/healthbox"
	"github.com/hotid/healthbox-hass-gw/pkg/homeassistant"
	"log"
	"sort"
	"strings"
	"time"
)

type GwDevices map[string]HaDevice

var ha *homeassistant.Client

type HaDevice struct {
	Name              string
	DeviceName        string
	UniqueId          string
	Identifiers       string
	State             string
	Unit              string
	AvailabilityTopic string
	StateTopic        string
	lastDiscovery     time.Time
	lastValue         time.Time
}

func (d *HaDevice) SetState(newState string) {
	d.State = newState
	d.lastValue = time.Now()
	d.PublishAvailability()
	d.PublishState()

}

func (d *HaDevice) PublishState() {
	token := ha.Mqtt.Publish(d.StateTopic, byte(0), false, d.State)
	token.Wait()
	if token.Error() != nil {
		fmt.Printf("%+v", d)
		panic(token.Error())
	}
}

func (d *HaDevice) PublishAvailability() {
	available := "0"
	if time.Since(d.lastValue) < time.Duration(60*time.Second) {
		available = "1"
	}
	token := ha.Mqtt.Publish(d.AvailabilityTopic, byte(0), false, available)
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}
}

func (d *HaDevice) PublishDiscovery() {
	payload := homeassistant.HaDeviceDiscoveryInfo{
		UniqueId:            d.UniqueId,
		ObjectId:            d.UniqueId,
		Name:                d.Name,
		PayloadAvailable:    "1",
		PayloadNotAvailable: "0",
		AvailabilityTopic:   d.AvailabilityTopic,
		StateTopic:          d.StateTopic,
		UnitOfMeasurement:   d.Unit,
	}
	payload.Device.Name = d.DeviceName
	payload.Device.Identifiers = "healthbox"
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("error marshalling payload: %s", err)
		return
	}
	topic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", "healthbox", d.UniqueId)
	log.Printf("publishing discivery data for device %s: %s %+v", d.Name, topic, string(jsonPayload))
	token := ha.Mqtt.Publish(topic, byte(0), false, jsonPayload)
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}
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
		devicesStr += fmt.Sprintf("%s: %s %s\n", device.Name, device.State, device.Unit)
	}
	return devicesStr
}

func (d GwDevices) Update(data *healthbox.CurrentData) {
	for _, room := range data.Room {
		state := fmt.Sprintf("%.1f", room.Actuator[0].Parameter.FlowRate.Value)
		deviceId := getUniqId(room)

		if device, ok := d[deviceId]; !ok {
			d[deviceId] = newHaDevice(room)
		} else {
			device.SetState(state)
		}
	}
}

func newHaDevice(room healthbox.RoomInfo) HaDevice {
	state := fmt.Sprintf("%.1f", room.Actuator[0].Parameter.FlowRate.Value)
	deviceId := getUniqId(room)

	device := HaDevice{
		Name:              room.Name,
		DeviceName:        deviceId,
		Identifiers:       "healthbox",
		UniqueId:          deviceId,
		State:             state,
		Unit:              room.Actuator[0].Parameter.FlowRate.Unit,
		AvailabilityTopic: fmt.Sprintf("%s/devices/controls/%s/availability", "healthbox", deviceId),
		StateTopic:        fmt.Sprintf("%s/devices/controls/%s", "healthbox", deviceId),
		lastValue:         time.Now(),
		lastDiscovery:     time.Now(),
	}
	device.PublishDiscovery()
	return device
}

func getUniqId(room healthbox.RoomInfo) string {
	name := strings.Replace(translit.EncodeToICAO(room.Name), " ", "_", -1)
	name = strings.Replace(name, "(", "", -1)
	name = strings.Replace(name, ")", "", -1)
	return fmt.Sprintf("%s_%s", name, "air_flow")
}
func (d GwDevices) StartDataUpdate(ctx context.Context, c *healthbox.Client) {
	go func() {
		for {
			currentData, err := c.GetCurrentData()
			if err != nil {
				log.Printf("error: %s", err)
				goto NextCycle
			}

			d.Update(currentData)
			log.Printf("%s", d)

		NextCycle:
			select {
			case <-ctx.Done():
				log.Printf("dataupdate task exiting")
				return
			default:
				time.Sleep(time.Second * 1)
			}
		}
	}()
}

func (d GwDevices) StartGateway(ctx context.Context, c *healthbox.Client, h *homeassistant.Client) {
	ha = h
	d.StartDataUpdate(ctx, c)
	d.StartDiscoveryPublishing(ctx)
}

func (d GwDevices) StartDiscoveryPublishing(ctx context.Context) {
	go func() {
		for {
			for _, device := range d {
				if time.Since(device.lastDiscovery) > time.Duration(3600*time.Second) {
					device.PublishDiscovery()
				}
			}
			select {
			case <-ctx.Done():
				log.Printf("discovery publishing task exiting")
				return
			default:
				time.Sleep(time.Second * 15)
			}
		}
	}()
}
