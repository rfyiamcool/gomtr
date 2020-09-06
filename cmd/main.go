package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/rfyiamcool/gomtr/mtr"
	"github.com/rfyiamcool/gomtr/spew"
)

var (
	count   = 3
	help    bool
	verbose bool
	ping    bool
	mtrFlag bool
	targets []string
)

func init() {
	flag.BoolVar(&help, "h", false, "print help()")
	flag.BoolVar(&verbose, "v", false, "verbose logging")
	flag.BoolVar(&mtrFlag, "mtr", true, "handle mtr")
	flag.BoolVar(&ping, "ping", false, "handle ping")
	flag.IntVar(&count, "c", 3, "run count")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stdout, `Usage: gomtr [-hvc] [-mtr] [-ping] hostname list

Options:
`)
	flag.PrintDefaults()
}

func parseCommand() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, errors.New("miss target params"))
		os.Exit(2)
	}
	targets = args
}

func main() {
	parseCommand()

	for _, addr := range targets {
		mm, err := mtr.Mtr(addr, 30, 3, 800)
		if err != nil {
			spew.Error(err)
		}
		spew.Debug(mm)
	}
}
