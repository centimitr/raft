package main

import "net"

func GetLocalIP() (ip string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addrs {
		ipn, ok := address.(*net.IPNet)
		if ok && ipn.IP.To4() != nil {
			ip = ipn.IP.String()
			if !ipn.IP.IsLoopback() {
				return
			}
		}
	}
	return
}
