// Copyright 2019 JD Inc. All Rights Reserved.
// type.go - file brief introduce
/*
modification history
----------------------------------------------
2019/5/26 0:20, by wangyulong3@jd.com, create

*/
/*
Description

*/

package mtr

import (
	"time"

	"github.com/rfyiamcool/gomtr/common"
)

const DEFAULT_MAX_HOPS = 30
const DEFAULT_TIMEOUT_MS = 800
const DEFAULT_PACKET_SIZE = 56
const DEFAULT_SNT_SIZE = 10

type MtrReturn struct {
	Success  bool
	TTL      int
	Host     string
	SuccSum  int
	LastTime time.Duration
	AllTime  time.Duration
	BestTime time.Duration
	AvgTime  time.Duration
	WrstTime time.Duration
}

type MtrResult struct {
	DestAddress string
	Hops        []common.IcmpHop
}

type MtrOptions struct {
	maxHops    int
	timeoutMs  int
	packetSize int
	sntSize    int
}

func (options *MtrOptions) MaxHops() int {
	if options.maxHops == 0 {
		options.maxHops = DEFAULT_MAX_HOPS
	}
	return options.maxHops
}

func (options *MtrOptions) SetMaxHops(maxHops int) {
	options.maxHops = maxHops
}

func (options *MtrOptions) TimeoutMs() int {
	if options.timeoutMs == 0 {
		options.timeoutMs = DEFAULT_TIMEOUT_MS
	}
	return options.timeoutMs
}

func (options *MtrOptions) SetTimeoutMs(timeoutMs int) {
	options.timeoutMs = timeoutMs
}

func (options *MtrOptions) SntSize() int {
	if options.sntSize == 0 {
		options.sntSize = DEFAULT_SNT_SIZE
	}
	return options.sntSize
}

func (options *MtrOptions) SetSntSize(sntSize int) {
	options.sntSize = sntSize
}

func (options *MtrOptions) PacketSize() int {
	if options.packetSize == 0 {
		options.packetSize = DEFAULT_PACKET_SIZE
	}
	return options.packetSize
}

func (options *MtrOptions) SetPacketSize(packetSize int) {
	options.packetSize = packetSize
}
