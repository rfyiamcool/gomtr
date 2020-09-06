package common

import (
	"log"
	"testing"

	"github.com/yinheli/qqwry"
)

func TestQueryIP(t *testing.T) {
	q := qqwry.NewQQwry("../qqwry.dat")
	q.Find("8.8.8.8")
	log.Printf("ip:%v, Country:%v, City:%v", q.Ip, q.Country, q.City)
}
