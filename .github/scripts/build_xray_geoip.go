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

		// Remove single quotes if wrapped
		line = strings.Trim(line, "'\"")

		// If it's just an IP without CIDR notation, append /32 or /128
		if !strings.Contains(line, "/") {
			if strings.Contains(line, ":") {
				line += "/128"
			} else {
				line += "/32"
			}
		}

		_, ipnet, err := net.ParseCIDR(line)
		if err != nil {
			log.Printf("Warning: skipping invalid entry: %s (%v)", line, err)
			continue
		}

		// Use the network IP from ipnet, and convert to proper byte length
		// V2Ray expects 4-byte IPs for IPv4 and 16-byte IPs for IPv6
		ip := ipnet.IP
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}

		ones, _ := ipnet.Mask.Size()

		cidrs = append(cidrs, &routercommon.CIDR{
			Ip:     ip,
			Prefix: uint32(ones),
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
