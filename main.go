package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/zlowram/gsd"
	"golang.org/x/net/proxy"
	"gopkg.in/mgo.v2"
)

const (
	TOP_100 = "7,9,13,21-23,25-26,37,53,79-81,88,106,110-111,113,119,135," +
		"139,143-144,179,199,389,427,443-445,465,513-515,543-544,548,554,587,631,646," +
		"873,990,993,995,1025-1029,1110,1433,1720,1723,1755,1900,2000-2001,2049,2121," +
		"2717,3000,3128,3306,3389,3986,4899,5000,5009,5051,5060,5101,5190,5357,5432," +
		"5631,5666,5800,5900,6000-6001,6646,7070,8000,8008-8009,8080-8081,8443,8888," +
		"9100,9999-10000,32768,49152-49157"
	ConfigFile     = "godan.toml"
	MAX_GOROUTINES = 100
)

var (
	file          = flag.Bool("f", false, "enable IP input from file")
	portsFlag     = flag.String("p", "", "set ports from cmdline (if not specified, top-100 ports)")
	portFile      = flag.String("P", "", "set ports from file (if not specified, top-100 ports)")
	maxGoroutines = flag.Int("m", MAX_GOROUTINES, "limit maximum number of goroutines")
	proxyHost     = flag.String("proxy", "", "set SOCKS5 proxy")
	proxyAuth     = flag.String("proxy-auth", "", "set authentication parameters for SOCKS5 proxy. Syntax: username:password")
)

func parseIPs() []string {
	var ips, ipList []string
	if *file {
		// Read list of IPs from file
		content, err := ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			log.Fatal(err)
		}

		ipList = strings.Split(string(content), "\n")
		ipList = ipList[:len(ipList)-1]
	} else {
		// Read IPs from comma-separated stuff
		ipList = strings.Split(flag.Arg(0), ",")
	}

	// Check if they are CIDR and expand
	for _, i := range ipList {
		cidrList := ipsFromCIDR(i)
		ips = append(ips, cidrList...)
	}

	return ips
}

func parsePorts() []string {
	var ports, portList []string
	if *portFile != "" {
		// Read list of IPs from file
		content, err := ioutil.ReadFile(*portFile)
		if err != nil {
			log.Fatal(err)
		}

		portList = strings.Split(string(content), "\n")
		portList = portList[:len(portList)-1]
	} else {
		// Read IPs from comma-separated stuff
		if *portsFlag == "" {
			*portsFlag = TOP_100
		}
		portList = strings.Split(*portsFlag, ",")
	}

	for _, i := range portList {
		if strings.Contains(i, "-") {
			sp := strings.Split(i, "-")
			prange, err := portRange(sp[0], sp[1])
			if err != nil {
				log.Fatal(err)
			}
			ports = append(ports, prange...)
		} else {
			ports = append(ports, i)
		}
	}

	return ports
}

func main() {
	// Parse arguments
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(2)
	}

	ips := parseIPs()
	ports := parsePorts()

	// Load the config
	config, err := loadConfig(flag.Arg(1))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stdout, "[+] Analyzing %d IPs and %d ports...\n", len(ips), len(ports))
	godan := gsd.NewGsd(ips, ports)

	// Add services
	services := []gsd.Service{
		//gsd.NewHttpsService(),
		//gsd.NewHttpService(),
		gsd.NewTCPService(),
		//gsd.NewTCPTLSService(),
	}
	godan.AddServices(services)

	// Check if proxy
	if *proxyHost != "" {
		var auth *proxy.Auth
		if *proxyAuth != "" {
			auth = &proxy.Auth{
				User:     strings.Split(*proxyAuth, ":")[0],
				Password: strings.Split(*proxyAuth, ":")[1],
			}
		} else {
			auth = &proxy.Auth{}
		}
		godan.SetProxy(*proxyHost, auth)
	}

	// Store the data!
	session, err := mgo.Dial(config.ServerIP + ":" + config.ServerPort)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	// Run them!
	results := godan.Run(*maxGoroutines)

	// Process results (filter errors out)
	for r := range results {
		if r.Error != "" {
			continue
		}
		c := session.DB(config.Database).C(config.Collection)
		err = c.Insert(&r)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("[+] DONE")
}

func usage() {
	fmt.Fprintln(os.Stderr, usageLine)
	flag.PrintDefaults()
}

const usageLine = `usage: godan [flags] <ips> <config>
  <ips>    if "-f" is not set: Comma-separated list of target IPs
           if "-f" is set: Path to file containing the list of target IPs
  <config> path to config file
flags:`
