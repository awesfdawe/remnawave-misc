package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

func main() {
	inputFile := flag.String("i", "ips.txt", "Input file with IPs/CIDRs")
	outputFile := flag.String("o", "geoip.dat", "Output geoip.dat file")
	flag.Parse()

	f, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}
	defer f.Close()

	var cidrs []*routercommon.CIDR
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Remove YAML list prefix if present
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") || line == "payload:" {
			continue
		}

		// Strip inline comments if any
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}

		// If it's just an IP without CIDR notation, append /32 or /128
		if !strings.Contains(line, "/") {
			if strings.Contains(line, ":") {
				line += "/128"
			} else {
				line += "/32"
			}
		}

		ip, ipnet, err := net.ParseCIDR(line)
		if err != nil {
			log.Printf("Warning: failed to parse %s: %v", line, err)
			continue
		}

		cidrs = append(cidrs, &routercommon.CIDR{
			Ip:     ip,
			Prefix: uint32(ones(ipnet.Mask)),
		})
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Read error: %v", err)
	}

	geoIP := &routercommon.GeoIP{
		CountryCode: "custom",
		Cidr:        cidrs,
	}
	geoIPList := &routercommon.GeoIPList{
		Entry: []*routercommon.GeoIP{geoIP},
	}

	data, err := proto.Marshal(geoIPList)
	if err != nil {
		log.Fatalf("Marshal error: %v", err)
	}

	err = os.WriteFile(*outputFile, data, 0644)
	if err != nil {
		log.Fatalf("Write error: %v", err)
	}

	log.Printf("Successfully compiled %d IPs/CIDRs to %s", len(cidrs), *outputFile)
}

func ones(mask net.IPMask) int {
	var count int
	for _, b := range mask {
		for i := 0; i < 8; i++ {
			if b&(1<<uint(7-i)) != 0 {
				count++
			}
		}
	}
	return count
}
