package mtr

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/rfyiamcool/gomtr/common"
	"github.com/rfyiamcool/gomtr/icmp"
	"github.com/rfyiamcool/gomtr/spew"
)

func Mtr(ipAddr string, maxHops, sntSize, timeoutMs int) (result string, err error) {
	options := MtrOptions{}
	options.SetMaxHops(maxHops)
	options.SetSntSize(sntSize)
	options.SetTimeoutMs(timeoutMs)

	var out MtrResult
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Start: %v, DestAddr: %v\n", time.Now().Format("2006-01-02 15:04:05"), ipAddr))
	out, err = runMtr(ipAddr, &options)
	if err != nil {
		buffer.WriteString(fmt.Sprintf("mtr failed due to an error: %v\n", err))
		return buffer.String(), err
	}
	if len(out.Hops) == 0 {
		buffer.WriteString("mtr failed. Expected at least one hop\n")
		return buffer.String(), nil
	}

	buffer.WriteString(fmt.Sprintf("%-3v  %-15v  %10v%c  %10v  %10v  %10v  %10v  %10v  %-100v \n", "", "HOST", "Loss", '%', "Snt", "Last", "Avg", "Best", "Wrst", "GEO"))

	var hopStr string
	var lastHop int
	for index, hop := range out.Hops {
		if hop.Success {
			if hopStr != "" {
				buffer.WriteString(hopStr)
				hopStr = ""
			}

			buffer.WriteString(fmt.Sprintf("%-3d  %-15v  %10.1f%c  %10v  %10.2f  %10.2f  %10.2f  %10.2f  %-100v \n", hop.TTL, hop.Address, hop.Loss, '%', hop.Snt, common.Time2Float(hop.LastTime), common.Time2Float(hop.AvgTime), common.Time2Float(hop.BestTime), common.Time2Float(hop.WrstTime), geoFor(hop.Address)))
			lastHop = hop.TTL
		} else {
			if index != len(out.Hops)-1 {
				hopStr += fmt.Sprintf("%-3d  %-15v  %10.1f%c  %10v  %10.2f  %10.2f  %10.2f  %10.2f  %-100v \n", hop.TTL, "???", float32(100), '%', int(0), float32(0), float32(0), float32(0), float32(0), "null")
			} else {
				lastHop++
				buffer.WriteString(fmt.Sprintf("%-3d %-48v\n", lastHop, "???"))
			}
		}
	}

	return buffer.String(), nil
}

func geoFor(ip string) string {
	country, city, err := common.GetIpInfo(ip)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%s:%s", country, city)
}

func getSafeIdent() int {
	thold := 1000
	gid := common.Goid()
	if gid < thold {
		return gid + thold
	}
	return gid
}

func batchDetect(destAddr string, options *MtrOptions) [][]*common.IcmpResp {
	var (
		defaultSeq = 0
		resps      = make([][]*common.IcmpResp, options.SntSize())
		timeout    = time.Duration(options.TimeoutMs()) * time.Millisecond

		lock sync.Mutex
	)

	for idx := range resps {
		resps[idx] = make([]*common.IcmpResp, options.MaxHops()+1)
	}

	for snt := 0; snt < options.SntSize(); snt++ {
		wg := sync.WaitGroup{}
		for ttl := 1; ttl < options.MaxHops(); ttl++ {
			ttl := ttl // copy
			wg.Add(1)
			go func() {
				defer wg.Done()

				// avoid ping resp conflict when multi gorouting handle ping
				pid := getSafeIdent()
				data, err := icmp.Icmp(destAddr, ttl, pid, timeout, defaultSeq)
				if err != nil || !data.Success {
					spew.Infof("failed to ping icmp, err: %s", err.Error())
					return
				}

				lock.Lock()
				resps[0][ttl] = &data
				lock.Unlock()
			}()
		}
		wg.Wait()
	}
	return resps
}

func runMtr(destAddr string, options *MtrOptions) (result MtrResult, err error) {
	result.Hops = []common.IcmpHop{}
	result.DestAddress = destAddr
	mtrResults := make([]*MtrReturn, options.MaxHops()+1) // not use first index

	multiResps := batchDetect(destAddr, options)
	for _, resps := range multiResps {
		for ttl := 1; ttl < options.MaxHops(); ttl++ {
			// init
			if mtrResults[ttl] == nil {
				mtrResults[ttl] = &MtrReturn{TTL: ttl, Host: "???", SuccSum: 0, Success: false, LastTime: time.Duration(0), AllTime: time.Duration(0), BestTime: time.Duration(0), WrstTime: time.Duration(0), AvgTime: time.Duration(0)}
			}

			// padding
			data := resps[ttl]
			if data == nil {
				continue
			}

			mtrResults[ttl].SuccSum = mtrResults[ttl].SuccSum + 1
			mtrResults[ttl].Host = data.Addr
			mtrResults[ttl].LastTime = data.Elapsed
			if mtrResults[ttl].WrstTime == time.Duration(0) || data.Elapsed > mtrResults[ttl].WrstTime {
				mtrResults[ttl].WrstTime = data.Elapsed
			}
			if mtrResults[ttl].BestTime == time.Duration(0) || data.Elapsed < mtrResults[ttl].BestTime {
				mtrResults[ttl].BestTime = data.Elapsed
			}
			mtrResults[ttl].AllTime += data.Elapsed
			mtrResults[ttl].AvgTime = time.Duration((int64)(mtrResults[ttl].AllTime/time.Microsecond)/(int64)(mtrResults[ttl].SuccSum)) * time.Microsecond
			mtrResults[ttl].Success = true

			if common.IsEqualIp(data.Addr, destAddr) {
				continue
			}
		}
	}

	// for snt := 0; snt < options.SntSize(); snt++ {
	// 	for ttl := 1; ttl < options.MaxHops(); ttl++ {
	// 		data, err := icmp.Icmp(destAddr, ttl, pid, timeout, seq)
	// 		if err != nil || !data.Success {
	// 			spew.Infof("failed to ping icmp, err: %s", err.Error())
	// 			continue
	// 		}

	// 		// init
	// 		if mtrResults[ttl] == nil {
	// 			mtrResults[ttl] = &MtrReturn{TTL: ttl, Host: "???", SuccSum: 0, Success: false, LastTime: time.Duration(0), AllTime: time.Duration(0), BestTime: time.Duration(0), WrstTime: time.Duration(0), AvgTime: time.Duration(0)}
	// 		}

	// 		// padding
	// 		mtrResults[ttl].SuccSum = mtrResults[ttl].SuccSum + 1
	// 		mtrResults[ttl].Host = data.Addr
	// 		mtrResults[ttl].LastTime = data.Elapsed
	// 		if mtrResults[ttl].WrstTime == time.Duration(0) || data.Elapsed > mtrResults[ttl].WrstTime {
	// 			mtrResults[ttl].WrstTime = data.Elapsed
	// 		}
	// 		if mtrResults[ttl].BestTime == time.Duration(0) || data.Elapsed < mtrResults[ttl].BestTime {
	// 			mtrResults[ttl].BestTime = data.Elapsed
	// 		}
	// 		mtrResults[ttl].AllTime += data.Elapsed
	// 		mtrResults[ttl].AvgTime = time.Duration((int64)(mtrResults[ttl].AllTime/time.Microsecond)/(int64)(mtrResults[ttl].SuccSum)) * time.Microsecond
	// 		mtrResults[ttl].Success = true

	// 		if common.IsEqualIp(data.Addr, destAddr) {
	// 			break
	// 		}
	// 	}
	// }

	for index, mtrResult := range mtrResults {
		if index == 0 {
			continue
		}

		if mtrResult == nil {
			break
		}

		hop := common.IcmpHop{TTL: mtrResult.TTL, Snt: options.SntSize()}
		hop.Address = mtrResult.Host
		hop.Host = mtrResult.Host
		hop.AvgTime = mtrResult.AvgTime
		hop.BestTime = mtrResult.BestTime
		hop.LastTime = mtrResult.LastTime
		failSum := options.SntSize() - mtrResult.SuccSum
		loss := (float32)(failSum) / (float32)(options.SntSize()) * 100
		hop.Loss = float32(loss)
		hop.WrstTime = mtrResult.WrstTime
		hop.Success = mtrResult.Success

		result.Hops = append(result.Hops, hop)

		if common.IsEqualIp(hop.Host, destAddr) {
			break
		}
	}

	return result, nil
}
