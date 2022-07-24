package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/config"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/gateway"
	"github.com/hotid/healthbox-hass-gw/internal/pkg/utils"
	"github.com/hotid/healthbox-hass-gw/pkg/healthbox"
	"github.com/hotid/healthbox-hass-gw/pkg/homeassistant"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfgPath := flag.String("config", "/etc/healthbox-hass-gw/healthbox-hass-gw.yaml", "config file location")
	flag.Parse()

	cfg, err := config.NewConfig(*cfgPath)
	if err != nil {
		panic(fmt.Sprintf("Error reading configuration file: %s", err))
	}
	var gw gateway.Gw

	healthboxClient := healthbox.NewClient(cfg)
	mqttClient := homeassistant.NewClient(cfg)
	go gw.StartGateway(ctx, healthboxClient, mqttClient)

	utils.HandleSignals(cancel)
	utils.WaitCancel(ctx)

}
