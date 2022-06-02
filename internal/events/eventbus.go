package events

import (
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/soerenschneider/dyndns/internal/common"
)

type EventDispatch interface {
	Notify(msg common.Envelope) error
}

type EventListener interface {
	Subscribe(func(client mqtt.Client, msg mqtt.Message)) error
}
