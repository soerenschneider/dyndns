package conf

import (
	"errors"
	"log"
)

type InterfaceConfig struct {
	NetworkInterface string `json:"interface"`
}

func (conf *InterfaceConfig) Print() {
	log.Printf("NetworkInterface=%s", conf.NetworkInterface)
}

func (conf *InterfaceConfig) Validate() error {
	if len(conf.NetworkInterface) == 0 {
		return errors.New("empty network interface provided")
	}

	return nil
}
