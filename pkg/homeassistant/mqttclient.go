package homeassistant

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"time"
)

type Client struct {
	Mqtt            mqtt.Client
	discoveryPrefix string
	payloadOnline   string
	payloadOffline  string
	component       string
	statusTopic     string
}

func (c *Client) PayloadOnline() string {
	return c.payloadOnline
}

func (c *Client) SetPayloadOnline(payloadOnline string) {
	c.payloadOnline = payloadOnline
}

func (c *Client) PayloadOffline() string {
	return c.payloadOffline
}

func (c *Client) SetPayloadOffline(payloadOffline string) {
	c.payloadOffline = payloadOffline
}

func (c *Client) StatusTopic() string {
	return c.statusTopic
}

func (c *Client) SetStatusTopic(statusTopic string) {
	c.statusTopic = statusTopic
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
		panic(token.Error())
	}

	c := &Client{Mqtt: mqttClient}
	c.SetDiscoveryPrefix(config.Mqtt.DiscoveryPrefix)
	c.SetPayloadOnline(config.Mqtt.StatusPayloadOnline)
	c.SetPayloadOffline(config.Mqtt.StatusPayloadOffline)
	c.SetStatusTopic(config.Mqtt.StatusTopic)

	return c
}

func (c *Client) SetDiscoveryPrefix(topic string) {
	c.discoveryPrefix = topic
}

func (c *Client) GetDiscoveryPrefix() string {
	return c.discoveryPrefix
}

func (c *Client) Publish(topic string, payload interface{}) {
	token := c.Mqtt.Publish(topic, byte(0), false, payload)
	token.Wait()
	if token.Error() != nil {
		panic(token.Error())
	}
}
