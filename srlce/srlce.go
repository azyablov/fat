package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/azyablov/fat/lib"
	"github.com/azyablov/fat/lib/gnoi/file"
	"github.com/azyablov/fat/lib/jrpc"
	"github.com/scrapli/scrapligo/driver/options"
	"github.com/scrapli/scrapligo/platform"
	"github.com/scrapli/scrapligo/response"
	"github.com/scrapli/scrapligo/util"
	log "github.com/sirupsen/logrus"
)

type cliOpt struct {
	printTree         *bool
	jsonRPC           *bool
	cleanUpClabConfig *bool
	gNOIdld           *bool
	rFile             *string
	logFile           *string
	d                 *bool
	logSSH            *bool
}

type showVersion map[string]string
type jsonOutText map[string]string

func main() {
	// Init and parse flags
	f := new(cliOpt)
	f.printTree = flag.Bool("printTree", false, "Print info object tree.")
	f.jsonRPC = flag.Bool("jsonrpc", false, "Use JSON RPC instead of SSH.")
	f.cleanUpClabConfig = flag.Bool("cclab", false, "Clean up clab generated config.")
	f.gNOIdld = flag.Bool("gNOIdld", false, "Use gNOI to download info config.")
	f.rFile = flag.String("rFile", "myconfig2.cfg", "Remote file name on target NE.")
	f.logFile = flag.String("logFile", "", "Log all messages into specified log file instead of stderr.")
	f.d = flag.Bool("d", false, "Enable debug, by default warn.")
	f.logSSH = flag.Bool("logSSH", false, "Enable SSH debug, s.")

	t := new(lib.SRLTarget)
	t.Username = flag.String("username", "admin", "SSH username")
	t.Password = flag.String("password", "admin", "SSH password")
	t.NoStrictKey = flag.Bool("noSKey", true, "No SSH key checking")
	t.Hostname = flag.String("target", "", "Target hostname")
	t.PortSSH = flag.Int("SSHport", 22, "SSH port")
	t.PortJRpc = flag.Int("JRPCport", 443, "JSON RPC port")
	t.PortgNOI = flag.Int("gNOIport", 57400, "gNOI port")
	t.Timeout = flag.Duration("timeout", 10*time.Second, "Connection timeout.")
	t.InsecConn = flag.Bool("InsecConn", false, "TLS insecure connectivity")
	t.SkipVerify = flag.Bool("SkipVerify", false, "skip TLS certificate chain verification")
	t.RootCA = flag.String("rootCA", "", "CA certificate file in PEM format.")
	t.Cert = flag.String("cert", "", "Client certificate file in PEM format.")
	t.Key = flag.String("key", "", "Client private key file.")
	flag.Parse()

	// init operational vars
	var (
		cfgFileName             string             // name of config file
		cfg                     string             // dev configuration
		cfgCClab                string             // dev config cleaned up from clab artifacts
		o                       *response.Response // SSH response
		okHostname, okSwVersion bool               // vars to verify key presence in order to build file name components
		hostname, swVersion     string
		cmdShVer                string = "show version"
		cmdInfo                 string = "info | as text"
		cmdInfoPipe             string = fmt.Sprintf("info > /tmp/%s", *f.rFile)
		cmdInfoPipeJRpc         string = fmt.Sprintf("info | as text > /tmp/%s", *f.rFile)
		cmdRmFilejRpc           string = fmt.Sprintf("file rm /tmp/%s", *f.rFile)
	)

	// Checking mandatory params
	switch {
	case len(*t.Hostname) == 0:
		log.WithFields(log.Fields{
			"exec": "checking flags and input params",
		}).Fatalln("hostname is mandatory, but missed")
	default:
	}

	// setup logging
	log.SetReportCaller(true)
	//// level
	log.SetLevel(log.WarnLevel)
	if *f.d {
		log.SetLevel(log.DebugLevel)
	}

	//// log destination config
	contextLogger := log.WithFields(log.Fields{
		"topic": "logging",
	})
	if *f.logFile != "" {
		contextLogger.Infof("File logging requested to setup into %s", *f.logFile)
		fh, err := os.OpenFile(*f.logFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0740)
		if err != nil {
			log.WithFields(log.Fields{
				"exec": "log setup",
			}).Errorf("can't create/open log file: %s", err)
		}
		defer fh.Close()
		log.SetOutput(fh)
	} else {
		contextLogger.Info("Using stdout for logging")
		log.SetOutput(os.Stderr)
	}

	if *f.jsonRPC {
		// Connecting via JSON-RPC
		contextLogger := log.WithFields(log.Fields{
			"topic": "JSON RPC",
		})
		contextLogger.Debug("Connecting via JSON-RPC...")
		var resp *jrpc.JSONRpcResponse
		var outShVer []map[string]showVersion
		// , exec sh ver
		resp, err := jrpc.ExecCli(t, &cmdShVer, jrpc.OutFormJSON)
		if err != nil {
			contextLogger.Fatalf("error while exec cli command: %s", err)
		}

		err = json.Unmarshal(lib.JSONRawToByte(resp.Result), &outShVer)
		contextLogger.Infof("resp.Result: %v", string(lib.JSONRawToByte(resp.Result)))
		if err != nil {
			contextLogger.Fatalf("[]byte marshalling error: %s", err)
		}

		hostname, okHostname = outShVer[0]["basic system info"]["Hostname"]
		swVersion, okSwVersion = outShVer[0]["basic system info"]["Software Version"]
		if !(okHostname && okSwVersion) {
			contextLogger.Fatalf("failed to parse show version output; error: %+v\n", err)
		}

		// , exec info / create file
		if *f.gNOIdld {
			// Creating file to download with gNOI
			contextLogger.Debug("Creating file to download with gNOI...")
			resp, err = jrpc.ExecCli(t, &cmdInfoPipeJRpc, jrpc.OutFormText)
			if err != nil {
				contextLogger.Fatalf("error while exec cli command: %s", err)
			}
			// check for the errors
			contextLogger.Infof("checking for command execution errors; resp.Result: %s", string(lib.JSONRawToByte(resp.Result)))
			if strings.Compare(string(lib.JSONRawToByte(resp.Result)), "[{}]") != 0 {
				contextLogger.Fatalf("expect no outputs, but got: %s", lib.JSONRawToByte(resp.Result))
			}
			*f.rFile = fmt.Sprintf("/tmp/%s", *f.rFile)
			// defer cleanup, bcz specific permissions jsonrpc:tls
			defer jrpc.ExecCli(t, &cmdRmFilejRpc, jrpc.OutFormText)

		} else {
			var outJSONText []jsonOutText
			var ok bool
			// exec info and scrape it
			contextLogger.Debugf("exec %s and scrape it", cmdInfo)
			resp, err = jrpc.ExecCli(t, &cmdInfo, jrpc.OutFormText)
			if err != nil {
				contextLogger.Debugf("JSON RPC response: ", resp)
				contextLogger.Fatalf("error while exec cli command: %s", err)
			}
			err = json.Unmarshal(lib.JSONRawToByte(resp.Result), &outJSONText)
			if err != nil {
				contextLogger.Fatalf("[]byte marshalling error: %s", err)
			}
			cfg, ok = outJSONText[0]["text"] // populating cfg info
			contextLogger.Infof("populating cfg info outJSONText[0]['text'] trunk up to 128: %s", outJSONText[0]["text"][:128])
			if !ok {
				log.WithFields(log.Fields{
					"exec": "JSON RPC",
				}).Fatalf("expected text element, but was not found")
			}
		}

	} else {
		// Connecting via SSH and issue necessary commands
		contextLogger := log.WithFields(log.Fields{
			"topic": "SSH",
		})
		contextLogger.Debug("Connecting via SSH and issue necessary commands...")
		var opts []util.Option
		if *t.NoStrictKey {
			opts = append(opts, options.WithAuthNoStrictKey())
		}
		opts = append(opts,
			options.WithAuthPassword(*t.Password),
			options.WithAuthUsername(*t.Username),
		)

		if *f.logSSH {
			opts = append(opts, options.WithDefaultLogger())
		}

		// Creating platform object and driver
		contextLogger.Debug("Creating platform object and driver")
		p, err := platform.NewPlatform(
			platform.NokiaSrl,
			*t.Hostname,
			opts...)
		if err != nil {
			contextLogger.Fatalf("failed to create platform; error: %+v\n", err)
			return
		}

		d, err := p.GetNetworkDriver()
		if err != nil {
			contextLogger.Fatalf("failed to open driver; error: %+v\n", err)
			return
		}

		// Open SSH connection toward specified target
		contextLogger.Debug("Open SSH connection toward specified target")
		err = d.Open()
		if err != nil {
			log.WithFields(log.Fields{
				"topic": "SSH",
			}).Fatalf("failed to open SSH connection toward %+v; error: %+v\n", d.Transport.GetHost(), err)
			return
		}
		defer d.Close()

		// Classic HA-HA
		contextLogger.Debug("Exec show version")
		o, err = d.SendCommand("show version")
		if err != nil {
			contextLogger.Fatalf("failed to send command; error: %+v\n ", err)
			return
		}

		shVerMap, err := o.TextFsmParse(`../tfsm/srl.show_version.template`)
		if err != nil {
			contextLogger.Fatalf("failed to parse show version output; error: %+v\n", err)
		}
		hostname, okHostname = shVerMap[0]["Hostname"].(string)
		swVersion, okSwVersion = shVerMap[0]["SoftwareVersion"].(string)
		contextLogger.Infof("shVerMap[0]['Hostname'], shVerMap[0]['SoftwareVersion']: %+v, %+v\n", shVerMap[0]["Hostname"].(string), shVerMap[0]["SoftwareVersion"].(string))
		if !(okHostname && okSwVersion) {
			contextLogger.Fatalf("failed to passe show version output; error: %+v\n", err)
		}

		// , exec info
		contextLogger.Debug("Exec info")
		if *f.gNOIdld {
			// Creating file to download via gNOI
			contextLogger.Debug("Creating file to download via gNOI")
			o, err := d.SendCommand(cmdInfoPipe)
			if err != nil {
				log.WithFields(log.Fields{
					"exec": "SSH",
				}).Fatalf("failed to send command; error: %+v\n ", err)
				return
			}

			// check for the errors
			contextLogger.Debugf("checking for command execution errors; o.Result: %v", o.Result)
			if len(o.Result) != 0 {
				if strings.Contains(o.Result, "Permission denied") {
					contextLogger.Fatalf("file permission issue, file could be created via jsonrpc previously: %s\n", o.Result)
				}
				contextLogger.Fatalf("expect no output, but got: %s\n", o.Result)
				return
			}
			*f.rFile = fmt.Sprintf("/tmp/%s", *f.rFile)

		} else {
			// scraping info
			contextLogger.Debug("scraping info")
			o, err = d.SendCommand("info")
			if err != nil {
				contextLogger.Fatalf("failed to send command; error: %+v\n ", err)
				return
			}
			contextLogger.Debugf("o.Result trunk up to 128: %v", o.Result[:128])
			cfg = o.Result // populating info
		}
	}

	// Crafting the name of the local config file.
	cfgFileName = strings.Join([]string{hostname, swVersion}, "_")
	cfgFileName = strings.Join([]string{cfgFileName, "cfg"}, ".")
	log.WithFields(log.Fields{
		"topic": "cfgFileName",
	}).Debugf("Crafting the name of the local config file: %s", cfgFileName)

	if *f.gNOIdld {
		// Downloading file using gNOI, if option is selected
		contextLogger := log.WithFields(log.Fields{
			"topic": "SSH",
		})
		contextLogger.Debug("Downloading file using gNOI, if option is selected")
		err := file.GetFile(t, f.rFile, &cfgFileName)
		if err != nil {
			contextLogger.Fatalf("can't get file (gNOI): %s", err)
		}
		// clean up the target in case created not via JSON RPC
		if !*f.jsonRPC {
			contextLogger.Debug("clean up the target in case created not via JSON RPC")
			defer file.RemoveFile(t, f.rFile)
		}

		if !(*f.cleanUpClabConfig || *f.printTree) {
			contextLogger.Debug("No clean-up or printTree requested, exiting...")
			return
		}
		fh, err := os.OpenFile(cfgFileName, os.O_RDWR, 0740)
		if err != nil {
			contextLogger.Fatalf("can't open downloaded configuration file: %v", err)
		}
		c, err := ioutil.ReadAll(fh)
		fh.Close()
		if err != nil {
			contextLogger.Fatalf("can't read saved configuration file: %v", err)
		}
		cfg = string(c)
	}

	// Creating info tree
	var root *lib.InfoObject
	var err error
	if *f.cleanUpClabConfig || *f.printTree {
		// Creating info tree
		log.WithFields(log.Fields{
			"topic": "InfoObject",
		}).Debug("Creating info tree")
		root, err = lib.NewInfoObject(cfg)
		if err != nil {
			log.WithFields(log.Fields{
				"exec": "parsing cfg -> info tree",
			}).Error(err)
		}
	}

	// Clean-up clab configuration artifacts
	if *f.cleanUpClabConfig {
		log.WithFields(log.Fields{
			"topic": "cleanUpClabConfig",
		}).Debug("Clean-up clab configuration artifacts")
		cfgCClab, err = lib.CleanUpClabInfoObjects(root, cfg)
		if err != nil {
			log.WithFields(log.Fields{
				"exec": "clean up clab cfg elem",
			}).Error(err)
		}
		cfg = cfgCClab
	}

	// Printing info object Tree
	if *f.printTree {
		lib.PrintInfObjTree(root)
	}

	// Saving target configuration
	log.WithFields(log.Fields{
		"topic": "saveTargetConfig",
	}).Debug("saving config to the file")
	err = saveTargetConfig(cfgFileName, cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"topic": "saving config",
		}).Errorf("saving target cfg: %v", err)
	}
}

func saveTargetConfig(cfgFileName string, cfg string) error {
	fh, err := os.OpenFile(cfgFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0740)
	if err != nil {
		return fmt.Errorf("can't save config file: %s", err)
	}
	defer fh.Close()
	fh.WriteString(cfg)
	err = fh.Sync()
	if err != nil {
		return fmt.Errorf("can't flush config file to disk: %s", err)
	}
	return nil
}
