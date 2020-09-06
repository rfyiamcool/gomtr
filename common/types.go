package common

import (
	"time"
)

type IcmpReturn struct {
	Success bool
	Addr    string
	Elapsed time.Duration
}

type IcmpHop struct {
	Success  bool
	Address  string
	Host     string
	N        int
	TTL      int
	Snt      int
	LastTime time.Duration
	AvgTime  time.Duration
	BestTime time.Duration
	WrstTime time.Duration
	Loss     float32
}
