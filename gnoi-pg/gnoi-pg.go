package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"log"
	"os"
	"strings"
	"time"

	"flag"

	"github.com/azyablov/gnmi-pg/gnmilib"
	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/grpc"
	// log "github.com/golang/glog" // TODO: check to implement, if relevant
)

var (
	// Certificate/key and rootca
	rootCA = flag.String("rootCA", "", "CA certificate file in PEM format.")
	cert   = flag.String("cert", "", "Client certificate file in PEM format.")
	key    = flag.String("key", "", "Client private key file.")

	// Credentials
	username = flag.String("username", "admin", "The username to authenticate against target.")
	password = flag.String("password", "admin", "The password to authenticate against target.")

	// gNMI target connectivity options
	targetHostname = flag.String("hostname", "", "The target hostname used to verify the hostname returned by TLS handshake.")
	targetAddr     = flag.String("addr", "", "The target address in the format of host[:port], by default port is 57400.")

	// Connection options
	insecConn  = flag.Bool("insecure", false, "Insecure connection.")
	skipVerify = flag.Bool("skip_verify", false, "Diable certificate validation during TLS session ramp-up.")
	timeout    = flag.Duration("timeout", 10*time.Second, "Connection timeout.")

	// File options
	rFile = flag.String("remoteFile", "", "Path to remote file.")
	lFile = flag.String("localFile", "", "Path to local file.")
)

const (
	targetPort = 57400
)

func main() {

	// Parsing flags
	flag.Parse()
	if len(*targetAddr) == 0 {
		flag.Usage()
		log.Fatalf("addr is mandatory to provide")
	}
	// Setting up grpc options
	dOpts, err := gnmilib.SetupGNMISecureTransport(
		gnmilib.TLSInit{
			InsecConn:      *insecConn,
			SkipVerify:     *skipVerify,
			TargetHostname: *targetHostname,
			RootCA:         *rootCA,
			Cert:           *cert,
			Key:            *key,
		})
	if err != nil {
		log.Fatal(err)
	}
	// Set up a connection to the server.
	var t string // target address to connect using grpc.Dial()
	if strings.Contains(*targetAddr, ":") {
		t = *targetAddr
	} else {
		t = fmt.Sprintf("%s:%v", *targetAddr, targetPort)
	}

	// Dialing and getting gRPC connection
	gRPCconn, err := grpc.Dial(t, *dOpts...)
	if err != nil {
		log.Fatalf("can't not connect to the host %s due to %s", t, err)
	}
	defer gRPCconn.Close()

	// Creating context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	// Attaching credentials to the context
	uc := gnmilib.UserCredentials{
		Username: *username,
		Password: *password,
	}
	// Populating credential in provided context
	ctx, err = gnmilib.PopulateMDCredentials(ctx, uc)
	if err != nil {
		log.Fatalln(err)
	}

	// Constructing file Get request
	fGetReq := new(file.GetRequest)
	// TODO: update flags!
	fGetReq.RemoteFile = *rFile

	fClient := file.NewFileClient(gRPCconn)
	fGetStream, err := fClient.Get(ctx, fGetReq)
	if err != nil {
		log.Fatalf("can't exec Get request: %s", err)
	}

	// Initializing buffer
	fileBuf := new(bytes.Buffer)
	fileBuf.Grow(65536)
	var mHash *types.HashType
	for {
		getResp, err := fGetStream.Recv()
		if err != nil {
			log.Fatalf("can't get GetResponse: %s", err)
		}
		bMessage := getResp.GetContents()
		if bMessage != nil {
			fileBuf.Write(bMessage)
			// if fileBuf.Len() > 10485760 {
			// 	fileBuf.Grow(65536)
			// 	f.Write(fileBuf.Bytes())
			// } else {
			// 	f.Write(fileBuf.Bytes())
			// }

			continue
		}

		mHash = getResp.GetHash()

		break

	}
	var cHash hash.Hash
	switch mHash.GetMethod() {
	case types.HashType_MD5:
		cHash = md5.New()
	case types.HashType_SHA256:
		cHash = sha256.New()
	case types.HashType_SHA512:
		cHash = sha512.New()
	case types.HashType_UNSPECIFIED:
		log.Fatalf("don't know how to handle HashType_UNSPECIFIED: %s", err)
	default:
		log.Fatalf("inappropriate specification hash method: %s", err)
	}

	hSum := cHash.Sum(fileBuf.Bytes())
	if bytes.Equal(hSum, mHash.Hash) {
		log.Fatalf("hash sum received is wrong or error happened during transmission: %s", err)
	}
	// Opening local file for writing
	f, err := os.Create(*lFile)
	if err != nil {
		log.Fatalf("can't create file: %s", err)
	}
	defer f.Close()

	f.Write(fileBuf.Bytes())

}
