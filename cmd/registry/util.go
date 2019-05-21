package main

import (
	"net"
)

func GetLocalIP() (s string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		ipn, ok := address.(*net.IPNet)
		if ok && ipn.IP.To4() != nil {
			ip := ipn.IP
			if ip.IsLoopback() {
				s = ip.String()
			}
			if ip.IsGlobalUnicast() {
				return ip.String()
			}
		}
	}
	return
}
