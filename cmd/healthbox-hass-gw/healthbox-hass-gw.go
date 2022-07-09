package main

import (
	"context"
	"fmt"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/gateway"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/utils"
	"github.com/hotid/healthbox-hass-gw/pkg/healthbox"
	"github.com/hotid/healthbox-hass-gw/pkg/homeassistant"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg, err := config.NewConfig("healthbox-hass-gw.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading configuration file: %s", err))
	}
	var devices gateway.GwDevices
	devices = make(map[string]*gateway.HaDevice)

	healthboxClient := healthbox.NewClient(cfg)
	mqttClient := homeassistant.NewClient(cfg)
	devices.StartGateway(ctx, healthboxClient, mqttClient)

	utils.HandleSignals(cancel)
	utils.WaitCancel(ctx)

}
