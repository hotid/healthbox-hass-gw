package homeassistant

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"log"
	"time"
)

type Client struct {
	Mqtt            mqtt.Client
	discoveryPrefix string
	component       string
}

//topic = self._discovery_prefix + '/' + component + '/' + node_id + '/' + object_id + '/config'

func NewClient(config *config.Config) *Client {
	mqttServer := fmt.Sprintf("%s:%s", config.Mqtt.Host, config.Mqtt.Port)
	opts := mqtt.NewClientOptions().AddBroker(mqttServer).SetClientID("healthbox-hass-gw")
	opts.SetKeepAlive(2 * time.Second)
	//opts.SetDefaultPublishHandler(f)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetUsername(config.Mqtt.Username)
	opts.SetPassword(config.Mqtt.Password)

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Panic(token.Error())
	}

	c := &Client{Mqtt: mqttClient}
	c.SetDiscoveryPrefix(config.Mqtt.DiscoveryPrefix)
	return c
}

func (c *Client) SetDiscoveryPrefix(topic string) {
	c.discoveryPrefix = topic
}

func (c *Client) GetDiscoveryPrefix() string {
	return c.discoveryPrefix
}
