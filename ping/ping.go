package ping

import (
	"bytes"
	"fmt"
	"time"

	"github.com/rfyiamcool/gomtr/common"
	"github.com/rfyiamcool/gomtr/icmp"
	"github.com/rfyiamcool/gomtr/spew"
)

func Ping(addr string, count, timeout, interval int) (result string, err error) {
	pingOptions := &PingOptions{}
	pingOptions.SetCount(count)
	pingOptions.SetTimeoutMs(timeout)
	pingOptions.SetIntervalMs(interval)

	ipAddrs, err := common.DestAddrs(addr)
	if err != nil || len(ipAddrs) == 0 {
		return
	}

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Start %v, PING %v (%v)\n", time.Now().Format("2006-01-02 15:04:05"), addr, ipAddrs[0]))
	begin := time.Now().UnixNano() / 1e6
	pingResp := runPing(ipAddrs[0], pingOptions)
	end := time.Now().UnixNano() / 1e6

	buffer.WriteString(fmt.Sprintf("%v packets transmitted, %v packet loss, time %vms\n", count, pingResp.dropRate, end-begin))
	buffer.WriteString(fmt.Sprintf("rtt min/avg/max = %v/%v/%v ms\n", common.Time2Float(pingResp.wrstTime), common.Time2Float(pingResp.avgTime), common.Time2Float(pingResp.bestTime)))

	buffer.WriteString(fmt.Sprintf("rtt min/avg/max = %v/%v/%v ms\n", pingResp.wrstTime.String(), pingResp.avgTime.String(), pingResp.bestTime.String()))

	result = buffer.String()

	return
}

func runPing(ipaddr string, option *PingOptions) (pingResp PingResp) {
	pingResp = PingResp{}
	pingResp.destAddr = ipaddr

	var (
		pid        = common.Goid()
		timeout    = time.Duration(option.TimeoutMs()) * time.Millisecond
		interval   = option.IntervalMs()
		pingResult = PingResult{}
	)

	seq := 0
	for cnt := 0; cnt < option.Count(); cnt++ {
		icmprt, err := icmp.Icmp(ipaddr, DEFAULT_TTL, pid, timeout, seq)
		if err != nil || !icmprt.Success || !common.IsEqualIp(ipaddr, icmprt.Addr) {
			spew.Errorf("failed to ping addr %s, err: %v", ipaddr, err)
			continue
		}

		pingResult.succSum++
		if pingResult.wrstTime == time.Duration(0) || icmprt.Elapsed > pingResult.wrstTime {
			pingResult.wrstTime = icmprt.Elapsed
		}
		if pingResult.bestTime == time.Duration(0) || icmprt.Elapsed < pingResult.bestTime {
			pingResult.bestTime = icmprt.Elapsed
		}
		pingResult.allTime += icmprt.Elapsed
		pingResult.avgTime = time.Duration((int64)(pingResult.allTime/time.Microsecond)/(int64)(pingResult.succSum)) * time.Microsecond
		pingResult.success = true

		seq++

		time.Sleep(time.Duration(interval) * time.Millisecond)
	}

	if !pingResult.success {
		pingResp.success = false
		pingResp.dropRate = 100.0
		return
	}

	pingResp.success = pingResult.success
	pingResp.dropRate = float64(option.Count()-pingResult.succSum) / float64(option.Count())
	pingResp.avgTime = pingResult.avgTime
	pingResp.bestTime = pingResult.bestTime
	pingResp.wrstTime = pingResult.wrstTime

	return
}
