package main

import (
	"context"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/gateway"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/utils"
	"github.com/hotid/healthbox-hass-gw/pkg/healthbox"
	"github.com/hotid/healthbox-hass-gw/pkg/homeassistant"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	config, err := config.NewConfig("/etc/healthbox-hass-gw/healthbox-hass-gw.yaml")
	if err != nil {
		log.Panic(err)
	}
	var devices gateway.GwDevices
	devices = make(map[string]gateway.HaDevice)

	hbClient := healthbox.NewClient(config)
	mqttClient := homeassistant.NewClient(config)
	devices.StartGateway(ctx, hbClient, mqttClient)

	utils.HandleSignals(cancel)
	utils.WaitCancel(ctx)

}
