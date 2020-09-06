package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rfyiamcool/gomtr/mtr"
	"github.com/rfyiamcool/gomtr/ping"
)

var (
	count   = 3
	targets = []string{"47.252.95.42", "47.75.18.65", "47.104.38.82", "47.88.73.1", "47.91.8.42"}
)

func main() {
	for _, addr := range targets {
		fmt.Println("start detect addr: ", addr)

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func(target string) {
			defer wg.Done()

			for i := 0; i < count; i++ {
				mm, err := mtr.Mtr(target, 30, 10, 800)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(mm)
			}
		}(addr)

		wg.Add(1)
		go func(target string) {
			defer wg.Done()

			for i := 0; i < count; i++ {
				mm, err := ping.Ping(target, 1, 1000, 1)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(mm)
				time.Sleep(1 * time.Second)
			}
		}(addr)

		wg.Wait()
	}

	select {}
}
