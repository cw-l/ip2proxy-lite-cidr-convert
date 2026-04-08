package main

import (
	"log"
	"github.com/cw-l/ip2proxy-lite-cidr-convert"
)

func main() {
	err := converter.GenerateExceptionData("./testdata/samples")
	if err != nil {
		log.Fatal(err)
	}
}