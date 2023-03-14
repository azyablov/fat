package file

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/azyablov/fat/lib"
	"github.com/azyablov/gnmi-pg/gnmilib"
	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/grpc"
)

func GetFile(t *lib.SRLTarget, rFile *string, lFile *string) error {
	r, err := GetReader(t, rFile)
	if err != nil {
		return err
	}

	// Opening local file for writing
	f, err := os.Create(*lFile)
	if err != nil {
		return fmt.Errorf("can't create file: %s", err)
	}
	defer f.Close()

	// pulling all data
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("can't read all from buffer: %s", err)
	}
	// writing to the file
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("unable to write data in file: %s", err)
	}
	return nil
}

func GetReader(t *lib.SRLTarget, rFile *string) (io.Reader, error) {

	// Setting up grpc options
	dOpts, err := gnmilib.SetupGNMISecureTransport(
		gnmilib.TLSInit{
			InsecConn:      *t.InsecConn,
			SkipVerify:     *t.SkipVerify,
			TargetHostname: *t.Hostname,
			RootCA:         *t.RootCA,
			Cert:           *t.Cert,
			Key:            *t.Key,
		})
	if err != nil {
		return nil, fmt.Errorf("setting up grpc options: %s", err)
	}
	// Set up a connection to the server.
	var host string // target address to connect using grpc.Dial()
	if strings.Contains(*t.Hostname, ":") {
		host = *t.Hostname
	} else {
		host = fmt.Sprintf("%s:%v", *t.Hostname, t.PortgNOI)
	}

	// Dialing and getting gRPC connection.
	gRPCconn, err := grpc.Dial(host, *dOpts...)
	if err != nil {
		return nil, fmt.Errorf("can't not connect to the host %s due to %s", *t.Hostname, err)
	}
	defer gRPCconn.Close()

	// Creating context with timeout.
	ctx, cancel := context.WithTimeout(context.Background(), *t.Timeout)
	defer cancel()

	// Attaching credentials to the context
	uc := gnmilib.UserCredentials{
		Username: *t.Username,
		Password: *t.Password,
	}
	// Populating credential in provided context
	ctx, err = gnmilib.PopulateMDCredentials(ctx, uc)
	if err != nil {
		return nil, fmt.Errorf("can't populate context with credentials: %s", err)
	}

	// Constructing file Get request
	fGetReq := new(file.GetRequest)
	fGetReq.RemoteFile = *rFile
	// Initializing buffer and hash var
	fileBuf := new(bytes.Buffer)
	fileBuf.Grow(65536)
	var mHash *types.HashType

	fClient := file.NewFileClient(gRPCconn)
	fGetStream, err := fClient.Get(ctx, fGetReq)
	if err != nil {
		return nil, fmt.Errorf("can't exec Get request: %s", err)
	}

	// Stream handling.
	for {
		getResp, err := fGetStream.Recv()
		if err != nil {
			return nil, fmt.Errorf("can't get GetResponse: %s", err)
		}
		// file contents
		bMessage := getResp.GetContents()
		if bMessage != nil {
			fileBuf.Write(bMessage)
			continue
		}
		// or hash
		mHash = getResp.GetHash()
		break

	}
	// Checking hash.
	var cHash hash.Hash
	switch mHash.GetMethod() {
	case types.HashType_MD5:
		cHash = md5.New()
	case types.HashType_SHA256:
		cHash = sha256.New()
	case types.HashType_SHA512:
		cHash = sha512.New()
	case types.HashType_UNSPECIFIED:
		return nil, fmt.Errorf("don't know how to handle HashType_UNSPECIFIED: %s", err)
	default:
		return nil, fmt.Errorf("inappropriate specification hash method: %s", err)
	}

	hSum := cHash.Sum(fileBuf.Bytes())
	if bytes.Equal(hSum, mHash.Hash) {
		return nil, fmt.Errorf("hash sum received is wrong or error happened during transmission: %s", err)
	}

	return fileBuf, nil

}
