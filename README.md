# Renson Healthbox 3.0 - Home Assistant geteway

Gateway to connect [Renson Healthbox 3.0](https://www.renson.eu/gd-gb/producten-zoeken/ventilatie/mechanische-ventilatie/units/healthbox-3-0) to [Home Assistant](http://home-assistant.io)
with [MQTT discovery](https://www.home-assistant.io/docs/mqtt/discovery).

* All rooms configured on Healthbox device should be added to HomeAssistant automatically.
* You don't need to write manually configuration for each sensor/switch/etc...
* You don't need manually configure MQTT bridge. All needed messages between Healthbox and Home Assistant will be forwarded by the gateway.

## Installation and configuration

### Requirements

* Go

### Installing:

```shell script
go install github.com/hotid/healthbox-hass-gw/cmd/healthbox-hass-gw
```
### Configuring healthbox-hass-gw

```yaml
Mqtt:
  Host: [YOUR MQTT BROKER HOST]
  Port: [YOUR MQTT BROKER PORT]
  Username: [YOUR MQTT USERNAME]
  Password: [YOUR MQTT PASSWORD]
Healthbox:
  Host: [YOUR HEALTHBOX HOST]
  Port: 8080
```

## Supported sensors:
* Flow rate

## Supported controls:
* Per room boost control

