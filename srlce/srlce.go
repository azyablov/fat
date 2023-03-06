package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/azyablov/fat/lib"
	"github.com/azyablov/fat/lib/jrpc"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/scrapli/scrapligo/response"
	"github.com/scrapli/scrapligo/util"
)

type authSSH struct {
	username    *string
	password    *string
	noStrictKey *bool
}

type targetHost struct {
	hostname *string
	//port     int
}

type tSRL struct {
	targetHost
	authSSH
}

type cliOpt struct {
	printTree         *bool
	jsonRPC           *bool
	cleanUpClabConfig *bool
}

func main() {
	// (0) Init and parse flags
	t := new(lib.SRLTarget)
	f := new(cliOpt)
	t.Username = flag.String("username", "admin", "SSH username")
	t.Password = flag.String("password", "admin", "SSH password")
	t.NoStrictKey = flag.Bool("noSKey", true, "No SSH key checking")
	f.printTree = flag.Bool("printTree", false, "Print info object tree.")
	f.jsonRPC = flag.Bool("jsonRPC", false, "Print info object tree.")
	f.cleanUpClabConfig = flag.Bool("cclab", false, "Clean up clab generated config.")

	t.Hostname = flag.String("target", "", "Target hostname")
	t.Port = flag.Int("port", 443, "Target port")
	flag.Parse()
	// Vars

	var (
		cfgFileName string // name of config file
		cfg         string // dev configuration
		cfgCClab    string // dev config cleaned up from clab artifacts
	)

	// Checking mandatory params

	switch {
	case len(*t.Hostname) == 0:
		// TODO: move to logrus
		log.Fatalln("hostname is mandatory, but missed")
	default:
	}

	var o *response.Response
	if *f.jsonRPC {
		// TODO: (1b) Connecting via JSON-RPC in case SSH failing or JSON-RPC.
		var resp *jrpc.JSONRpcResponse
		resp, err := jrpc.ExecCli(t, "show version")
		if err != nil {
			log.Fatalf("error while exec cli command: %s", err)
		}
		bJSON, err := resp.Result.MarshalJSON()
		if err != nil {
			log.Fatalf("unmarshalling error: %s", err)
		}
		fmt.Println(string(bJSON))
		return
		// TODO: (2a) Downloading file using gNOI
	} else {
		// (1a) Connecting via SSH and issue necessary commands
		log.Println("(1a) Connecting via SSH and issue necessary commands")

		var opts []util.Option
		if *t.NoStrictKey {
			opts = append(opts, options.WithAuthNoStrictKey())
		}
		opts = append(opts,
			// Add logger
			// 		func WithLogger ¶
			// func WithLogger(l *logging.Instance) util.Option
			// WithLogger accepts a logging.Instance and applies it to the driver object.
			options.WithAuthPassword(*t.Password),
			options.WithAuthUsername(*t.Username))

		// Creating platform object
		p, err := platform.NewPlatform(
			platform.NokiaSrl,
			*t.Hostname,
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

		// Open SSH connection toward specified target
		err = d.Open()
		if err != nil {
			fmt.Printf("failed to open SSH connection toward %+v; error: %+v\n", d.Transport.GetHost(), err)
			return
		}
		// Closing connection
		defer d.Close()

		o, err = d.SendCommand("show version")
		if err != nil {
			fmt.Printf("failed to send command; error: %+v\n ", err)
			return
		}
		fmt.Println(o.Result)
		shVerMap, err := o.TextFsmParse(`../tfsm/srl.show_version.template`)
		if err != nil {
			fmt.Printf("failed to passe show version output; error: %+v\n", err)
		}
		var okHostname, okSwVersion bool
		var hostname, swVersion string
		hostname, okHostname = shVerMap[0]["Hostname"].(string)
		swVersion, okSwVersion = shVerMap[0]["SoftwareVersion"].(string)
		if !(okHostname && okSwVersion) {
			fmt.Printf("failed to passe show version output; error: %+v\n", err)
		}
		// Crafting the name of local config file.
		cfgFileName = strings.Join([]string{hostname, swVersion}, "_")
		fmt.Printf("%+v\n", cfgFileName)

		o, err = d.SendCommand("info")
		if err != nil {
			fmt.Printf("failed to send command; error: %+v\n ", err)
			return
		}
		cfg = o.Result
	}

	// TODO: (3) Building info tree.
	fmt.Println("\nParsing....")

	//fmt.Println(o.Result)
	root, err := lib.NewInfoObject(cfg)
	if err != nil {
		log.Panicln(err)
	}

	// (4) Clean-up clab configuration artifacts
	if *f.cleanUpClabConfig {
		cfgCClab, err = lib.CleanUpClabInfoObjects(root, cfg)
		if err != nil {
			log.Panicln(err)
		}
		cfg = cfgCClab
	}

	// (5) Saving target configuration
	fh, err := os.OpenFile(cfgFileName, os.O_CREATE|os.O_RDWR, 0740)
	if err != nil {
		log.Fatalf("can't save config file: %v", err)
	}
	defer fh.Close()
	fh.WriteString(cfg)
	err = fh.Sync()
	if err != nil {
		log.Fatalf("can't flush config file to disk: %v", err)
	}

	// (6) Printing info object Tree
	if *f.printTree && root != nil {
		lib.PrintInfObjTree(root)
	}
}
