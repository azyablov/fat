package main

import (
	"flag"
	"log"

	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/scrapli/scrapligo/util"
)

type tSRL struct {
	username    *string
	password    *string
	noStrictKey *bool
}

func main() {
	// (0) Init and parse flags
	t := new(tSRL)
	t.username = flag.String("username", "admin", "SSH username")
	t.password = flag.String("password", "admin", "SSH password")
	t.noStrictKey = flag.Bool("noSKey", true, "No SSH key checking")
	flag.Parse()
	// (1a) Connecting via SSH and issue necessary commands
	// TODO: move to sirupsen/logrus
	log.Println("(1a) Connecting via SSH and issue necessary commands")
	var opts []util.Option
	if *t.noStrictKey {
		opts = append(opts, options.WithAuthNoStrictKey())
	}
	opts = append(opts,
		// Add logger
		// 		func WithLogger ¶
		// func WithLogger(l *logging.Instance) util.Option
		// WithLogger accepts a logging.Instance and applies it to the driver object.
		options.WithAuthPassword(*t.password),
		options.WithAuthUsername(*t.username))

	d, err := platform.NewPlatform(
		platform.NokiaSrl,
		opts...)

	// TODO: (1b) Connecting via JSON-RPC in case SSH failing or JSON-RPC more preferable

	// TODO: (2a) Downloading file using gNOI

	// TODO: (3) Clean-up clab configuration artifacts

	// TODO: (4) Saving target configuration

}
