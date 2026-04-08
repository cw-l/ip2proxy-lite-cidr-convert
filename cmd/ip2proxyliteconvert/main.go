package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net/netip"
	"os"
	"strings"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"go4.org/netipx"
)

type Schema struct {
	ColCount int
	Mapping  map[string]int
}

// version is injected by GoReleaser ldflags
var version = "dev" // Default value for local builds

var Registry = map[string]Schema{
	"px1":  {3, map[string]int{"country_code": 1, "country_name": 2}},
	"px2":  {4, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3}},
	"px3":  {6, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5}},
	"px4":  {7, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6}},
	"px5":  {8, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7}},
	"px6":  {9, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8}},
	"px7":  {11, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10}},
	"px8":  {12, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10, "last_seen": 11}},
	"px9":  {13, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10, "last_seen": 11, "threat": 12}},
	"px10": {13, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10, "last_seen": 11, "threat": 12}},
	"px11": {14, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10, "last_seen": 11, "threat": 12, "provider": 13}},
	"px12": {15, map[string]int{"proxy_type": 1, "country_code": 2, "country_name": 3, "region_name": 4, "city_name": 5, "isp": 6, "domain": 7, "usage_type": 8, "asn": 9, "as": 10, "last_seen": 11, "threat": 12, "provider": 13, "fraud_score": 14}},
}

func main() {
	in := flag.String("in", "", "Input CSV file")
	db := flag.String("db", "", "db (px1 to px12)")
	out := flag.String("out", "", "Output file (default: IP2PROXY-LITE-<PXN>.MMDB)")
	printVersion := flag.Bool("v", false, "Print version information")
	flag.Parse()

	if *printVersion {
		fmt.Printf("ip2proxyliteconvert version %s\n", version)
		os.Exit(0)
	}

	dbKey := strings.ToUpper(*db)
	conf, ok := Registry[strings.ToLower(*db)]

	// Check if input file is missing
	if *in == "" {
		log.Fatal("[FATAL] Missing input file. Use --in <file>")
	}

	// Check if the DB level is valid
	if !ok {
		log.Fatalf("[FATAL] Invalid DB level: '%s'. Must be px1 through px12.", *db)
	}

	// Determine the final filename
	finalOut := *out
	if finalOut == "" {
		finalOut = fmt.Sprintf("IP2PROXY-LITE-%s.MMDB", dbKey)
	}

	// Memory Optimisation: Create keys ONCE before the loop
	encodedKeys := make(map[string]mmdbtype.String)
	for field := range conf.Mapping {
		encodedKeys[field] = mmdbtype.String(field)
	}

	f, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.LazyQuotes = true

	writer, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "IP2Proxy-LITE-CIDR",
		RecordSize:   32,
		IPVersion:    6,
	})

	if err != nil {
		log.Fatalf("Writer init failed: %v", err)
	}

	var count, skipped, lineNum int
	for {
		lineNum++
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[LINE %d] CSV Read Error: %v", lineNum, err)
			continue
		}

		if len(rec) != conf.ColCount {
			skipped++
			if skipped > 100 && count == 0 {
				log.Fatalf("[FATAL] Massive Schema Mismatch. Check if --db %s is correct.", *db)
			}
			continue
		}

		prefix, err := netip.ParsePrefix(rec[0])
		if err != nil {
			skipped++
			continue
		}

		// Unwrap ::ffff:x.x.x.x into native IPv4
		if prefix.Addr().Is4In6() {
			addr := prefix.Addr().Unmap()
			prefix = netip.PrefixFrom(addr, prefix.Bits()-96)
		}

		// Normalise host bits
		prefix = prefix.Masked()

		if prefix.Addr().Is4() && prefix.Addr().IsPrivate() {
			skipped++
			continue
		}

		if prefix.Addr().Is6() && (prefix.Addr().IsPrivate() || prefix.Addr().IsLoopback() || prefix.Addr().IsLinkLocalUnicast() || prefix.Addr().IsUnspecified()) {
			skipped++
			continue
		}

		data := mmdbtype.Map{}
		for field, idx := range conf.Mapping {
			val := strings.TrimSpace(rec[idx])
			if val == "-" || val == "" {
				continue
			}
			data[encodedKeys[field]] = mmdbtype.String(val)
		}

		// Insert with Error Check
		err = writer.Insert(netipx.PrefixIPNet(prefix), data)
		if err != nil {
			skipped++
			continue
		}
		count++
	}

	if count == 0 {
		log.Fatal("[FATAL] No records inserted. Aborting — output file not created. Check source file.")
	}

	outFile, err := os.Create(finalOut)
	if err != nil {
		log.Fatalf("[FATAL] Could not create %s: %v", finalOut, err)
	}

	n, err := writer.WriteTo(outFile)
	if err != nil {
		log.Fatalf("[FATAL] WriteTo failed: %v", err)
	}

	// Flush to disk
	_ = outFile.Sync()
	outFile.Close()

	log.Printf("[SUCCESS] Generated %s (%d bytes) | Records: %d", finalOut, n, count)
}
