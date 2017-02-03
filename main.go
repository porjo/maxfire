package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"

	"github.com/oschwald/maxminddb-golang"
)

func main() {

	countryCode := flag.String("countryCode", "", "ISO Country code e.g. AU for Australia")
	ipv4 := flag.Bool("ipv4", true, "dump IPv4")
	ipv6 := flag.Bool("ipv6", false, "dump IPv6")
	fileName := flag.String("fileName", "", "GeoIP2 database file")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")

	flag.Parse()

	if *countryCode == "" || *fileName == "" {
		if *countryCode == "" {
			fmt.Println("Please provide a country code")
		}
		if *fileName == "" {
			fmt.Println("Please provide a file name")
		}
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	db, err := maxminddb.Open(*fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var record struct {
		Country struct {
			ISOCode string `maxminddb:"iso_code"`
		} `maxminddb:"country"`
	}

	networks := db.Networks()
	subnets := make([]*net.IPNet, 0)
	for networks.Next() {
		subnet, err := networks.Network(&record)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if subnet.IP.To4() != nil {
			if !*ipv4 {
				continue
			}
		} else if !*ipv6 {
			continue
		}

		if record.Country.ISOCode == *countryCode {
			subnets = append(subnets, subnet)
		}
	}
	if networks.Err() != nil {
		fmt.Println(networks.Err())
		os.Exit(1)
	}

	for _, subnet := range subnets {
		fmt.Printf("%s\n", subnet)
	}
}
