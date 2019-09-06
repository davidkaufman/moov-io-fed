// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

// fedtest is a cli tool for testing Moov's FED API endpoints.
//
// fedtest is not a stable tool. Please contact Moov developers if you intend to use this tool,
// otherwise we might change the tool (or remove it) without notice.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/fed"
	client "github.com/moov-io/fed/client"
)

var (
	flagLocal = flag.Bool("local", false, "Use local HTTP addresses (e.g. 'go run')")

	flagRoutingNumber = flag.String("routing-number", chaseCaliforniaRouting, "Routing number to lookup in FED")
)

const (
	chaseCaliforniaRouting = "322271627"
)

func main() {
	flag.Parse()

	log.SetFlags(log.Ldate | log.Ltime | log.LUTC | log.Lmicroseconds | log.Lshortfile)
	log.Printf("Starting fedtest %s", fed.Version)

	// setup our APIClient and ping
	api := client.NewAPIClient(makeConfig())
	if err := ping(api); err != nil {
		log.Fatalf("[FAILURE] %v", err)
	}

	requestID, routingNumber := base.ID(), *flagRoutingNumber
	log.Printf("[INFO] using x-request-id: %s", requestID)

	// ACH search
	if err := achSearch(api, requestID, routingNumber); err != nil {
		log.Fatalf("[FAILURE] ACH: error looking up %s: %v", routingNumber, err)
	} else {
		log.Printf("[SUCCESS] ACH: found %s", routingNumber)
	}

	// WIRE search
	if err := wireSearch(api, requestID, routingNumber); err != nil {
		log.Fatalf("[FAILURE] Wire: error looking up %s: %v", routingNumber, err)
	} else {
		log.Printf("[SUCCESS] Wire: found %s", routingNumber)
	}
}

func makeConfig() *client.Configuration {
	conf := client.NewConfiguration()
	if *flagLocal {
		conf.BasePath = "http://localhost"
	}
	conf.UserAgent = fmt.Sprintf("moov fedtest/%s", fed.Version)
	conf.HTTPClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			MaxConnsPerHost:     100,
			IdleConnTimeout:     1 * time.Minute,
		},
	}
	return conf
}

func ping(api *client.APIClient) error {
	resp, err := api.FEDApi.Ping(context.Background())
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()
	if err != nil {
		return fmt.Errorf("ping error: %v", err)
	}
	return nil
}
