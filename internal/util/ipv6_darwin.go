package util

import (
	"errors"
	"net/netip"
)

func LocalGlobalIpV6Addr(_ string) (netip.Addr, error) {
	return netip.Addr{}, errors.New("unsupported on darwin")
}
