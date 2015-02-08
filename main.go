package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/zlowram/gsd"
)

var (
	ipFile   = flag.Bool("i", false, "")
	portFile = flag.Bool("p", false, "")
)

func parseArgs() ([]string, []string) {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}

	var ips, ports []string
	if *ipFile {
		// Read list of IPs from file
		content, err := ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}

		ips = strings.Split(string(content), "\n")
		ips = ips[:len(ips)-1]
	} else {
		// Read IPs from comma-separated stuff
		ips = strings.Split(flag.Arg(0), ",")
	}

	if *portFile {
		// Read list of IPs from file
		content, err := ioutil.ReadFile(flag.Arg(1))
		if err != nil {
			log.Fatal(err)
		}

		ports = strings.Split(string(content), "\n")
		ports = ports[:len(ports)-1]
	} else {
		// Read IPs from comma-separated stuff
		ports = strings.Split(flag.Arg(1), ",")
	}

	return ips, ports
}

func main() {
	// Parse arguments
	ips, ports := parseArgs()

	godan := gsd.NewGsd(ips, ports)

	// Add services
	services := []gsd.Service{
		gsd.NewHttpsService(),
		gsd.NewHttpService(),
		gsd.NewTCPService(),
		gsd.NewTCPTLSService(),
	}
	godan.AddServices(services)

	// Run them!
	results := godan.Run()

	// Process results (filter errors out)
	for _, r := range results {
		if r.Error == "" {
			b, err := json.Marshal(r)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(b))
		}
	}
}

func usage() {
	usageline := `
	Flags:
		-i		Enable IP input from file.
		-p		Enable Port input from file.

	Args:
		<ips>		If "-i" is not set: Comma-separated list of target IPs.
				If "-i" is set: Path to file containing the list of target IPs.

		<ports>		If "-p" is not set: Comma-separated list of target ports.
				If "-p" is set: Path to file containin the list of target ports.
	`
	name := strings.Split(os.Args[0], "/")
	fmt.Fprintf(os.Stderr, "usage: %s [<flags>] <ips> <ports>%s", name[len(name)-1], usageline)
}
