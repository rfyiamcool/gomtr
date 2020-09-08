package common

import (
	"errors"
	"os"
	"sync"

	"github.com/yinheli/qqwry"
)

var (
	wryLock     sync.Mutex
	wrydataPath = "~/gomtr/qqwry.dat"
)

func GetIpInfo(ip string) (string, string, error) {
	wryLock.Lock()
	defer wryLock.Unlock()

	_, err := os.Stat(wrydataPath)
	if os.IsNotExist(err) {
		return "", "", errors.New("file qwray.dat not found")
	}

	// not thread safe
	q := qqwry.NewQQwry(wrydataPath)
	q.Find(ip)
	return q.Country, q.City, nil
}
