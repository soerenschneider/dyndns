package events

import (
	"github.com/soerenschneider/dyndns/internal/common"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type EventDispatch interface {
	Notify(msg common.Envelope) error
}

type EventListener interface {
	Subscribe(func(client mqtt.Client, msg mqtt.Message)) error
}
