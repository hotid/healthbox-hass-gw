# Renson Healthbox 3.0 - Home Assistant geteway

Gateway to connect [Renson Healthbox 3.0](https://www.renson.eu/gd-gb/producten-zoeken/ventilatie/mechanische-ventilatie/units/healthbox-3-0) to [Home Assistant](http://home-assistant.io)
with [MQTT discovery](https://www.home-assistant.io/docs/mqtt/discovery).

* All rooms configured on Healthbox device should be added to HomeAssistant automatically.
* You don't need to write manually configuration for each sensor/switch/etc...
* You don't need manually configure MQTT bridge. All needed messages between Healthbox and Home Assistant will be forwarded by the gateway.

## Installation and configuration

### Requirements

* Go
* Healthbox 3.0 & HomeAssistant ;)

### Installing as a service:

* Create configuration file (/etc/healthbox-hass-gw/healthbox-hass-gw.yaml by default)
* Create systemd unit file (/etc/systemd/system/healthbox-hass-gw.service)

```shell script
go install github.com/hotid/healthbox-hass-gw/cmd/healthbox-hass-gw
cp ~/go/bin/healthbox-hass-gw /usr/local/bin
systemctl daemon-reload
systemctl start healthbox-hass-gw
```

Observe syslog for errors and messages.

### Configuring healthbox-hass-gw

```yaml
Mqtt:
  Host: [YOUR MQTT BROKER HOST]
  Port: [YOUR MQTT BROKER PORT]
  Username: [YOUR MQTT USERNAME]
  Password: [YOUR MQTT PASSWORD]
  StatusTopic: 'hass/status'
  StatusPayloadOnline: 'online'
  StatusPayloadOffline: 'offline'
Healthbox:
  Host: [YOUR HEALTHBOX HOST]
  Port: 8080
```

## Supported sensors:
* Flow rate

## Supported controls:
* Per room boost control

## Example systemd unit file
```unit file (systemd)
[Unit]
Description="Healthbox to HomeAssistant mqtt gateway"

[Service]
User=nobody
Group=nogroup
ExecStart=/usr/local/bin/healthbox-hass-gw
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```
