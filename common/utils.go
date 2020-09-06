package common

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rfyiamcool/gomtr/spew"
)

// LookupAddrs nslookup domain name, return ips
func LookupIps(addr string) ([]string, error) {
	addrs, err := net.LookupHost(addr)
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0)
	for _, addr := range addrs {
		ipaddr, err := net.ResolveIPAddr("ip", addr)
		if err != nil {
			continue
		}

		ips = append(ips, ipaddr.IP.String())
	}

	return ips, nil
}

func Goid() int {
	defer func() {
		if err := recover(); err != nil {
			spew.Errorf("panic recover, err: %v", err)
		}
	}()

	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}

func IsEqualIp(ips1, ips2 string) bool {
	ip1 := net.ParseIP(ips1)
	if ip1 == nil {
		return false
	}

	ip2 := net.ParseIP(ips2)
	if ip2 == nil {
		return false
	}

	if ip1.String() != ip2.String() {
		return false
	}

	return true
}

func Time2Float(t time.Duration) float32 {
	return (float32)(t/time.Microsecond) / float32(1000)
}
