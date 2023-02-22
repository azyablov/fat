package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/scrapli/scrapligo/util"
)

type authSSH struct {
	username    *string
	password    *string
	noStrictKey *bool
}

type targetHost struct {
	hostname *string
	port     int
}

type tSRL struct {
	targetHost
	authSSH
}

func main() {
	// (0) Init and parse flags
	t := new(tSRL)
	t.username = flag.String("username", "admin", "SSH username")
	t.password = flag.String("password", "admin", "SSH password")
	t.noStrictKey = flag.Bool("noSKey", true, "No SSH key checking")

	t.hostname = flag.String("target", "", "Target hostname")
	flag.Parse()

	// Checking mandatory params

	switch {
	case len(*t.hostname) == 0:
		log.Fatalln("hostname is mandatory, but missed")
	default:
	}
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

	// Creating platform object
	p, err := platform.NewPlatform(
		platform.NokiaSrl,
		*t.hostname,
		opts...)
	if err != nil {
		fmt.Printf("failed to create platform; error: %+v\n", err)
		return
	}

	d, err := p.GetNetworkDriver()
	if err != nil {
		fmt.Printf("failed to open driver; error: %+v\n", err)
		return
	}

	err = d.Open()
	if err != nil {
		fmt.Printf("failed to open SSH connection toward %+v; error: %+v\n", d.Transport.GetHost(), err)
		return
	}
	// Closing connection
	defer d.Close()

	o, err := d.SendCommand("show version")
	if err != nil {
		fmt.Printf("failed to send command; error: %+v\n ", err)
		return
	}
	fmt.Println(o.Result)

	o, err = d.SendCommand("show platform chassis")
	if err != nil {
		fmt.Printf("failed to send command; error: %+v\n ", err)
		return
	}
	fmt.Println(o.Result)

	o, err = d.SendCommand("info")
	if err != nil {
		fmt.Printf("failed to send command; error: %+v\n ", err)
		return
	}
	fmt.Println(o.Result)

	// TODO: (1b) Connecting via JSON-RPC in case SSH failing or JSON-RPC more preferable

	// TODO: (2a) Downloading file using gNOI

	// TODO: (3) Clean-up clab configuration artifacts

	// TODO: (4) Saving target configuration

}
