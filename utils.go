package main

import (
	"errors"
	"net"
	"strconv"
)

func ipsFromCIDR(cidr string) []string {
	var ips []string

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		// Probably not a CIDR
		ips = append(ips, cidr)
		return ips
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func portRange(a string, b string) ([]string, error) {
	var ports []string

	n, _ := strconv.Atoi(a)
	m, _ := strconv.Atoi(b)

	if n >= m {
		return ports, errors.New("First parameter cannot be equal or greater than second")
	}

	for i := n; i <= m; i++ {
		ports = append(ports, strconv.Itoa(i))
	}

	return ports, nil
}
