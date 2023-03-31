package common

import (
	"os"
	"path"
	"sync"

	"github.com/mitchellh/go-homedir"
	"github.com/yinheli/qqwry"
)

var (
	home, _     = homedir.Dir()
	wrydataPath = path.Join(home, "gomtr/qqwry.dat")

	mu          sync.Mutex
	geoipParser = buildGeoipParser()
)

func GetIpInfo(ip string) (string, string, error) {
	if geoipParser == nil {
		return "", "", nil
	}

	// qqwry don't support thread safe.
	mu.Lock()
	defer mu.Unlock()

	geoipParser.Find(ip)
	return geoipParser.Country, geoipParser.City, nil
}

func buildGeoipParser() *qqwry.QQwry {
	_, err := os.Stat(wrydataPath)
	if os.IsNotExist(err) {
		return nil
	}

	return qqwry.NewQQwry(wrydataPath)
}
