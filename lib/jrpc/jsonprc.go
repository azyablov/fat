package jrpc

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/azyablov/fat/lib"
)

type Method string
type OutputFormat string

const (
	MethodGet    Method       = "get"
	MethodSet    Method       = "set"
	MethodCli    Method       = "cli"
	OutFormJSON  OutputFormat = "json"
	OutFormText  OutputFormat = "text"
	OutFormTable OutputFormat = "table"
)

// Request definition
type JSONRpcRequest struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	Method         Method `json:"method"`
	Params         Params `json:"params"`
}

// Response definition
type JSONRpcResponse struct {
	JSONRpcVersion string           `json:"jsonrpc"`
	Result         *json.RawMessage `json:"result,omitempty"`
	Error          *RpcError        `json:"error,omitempty"`
	ID             int              `json:"id"`
}

type Params struct {
	Commands  []interface{} `json:"commands"`
	OutFormat OutputFormat  `json:"output-format,omitempty"`
}

type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ExecCli(t *lib.SRLTarget, cmd *string, f OutputFormat) (*JSONRpcResponse, error) {

	if len(*cmd) == 0 || cmd == nil {
		return nil, fmt.Errorf("command can't be null string or nil")
	}
	var outFormat OutputFormat

	switch f {
	case OutFormJSON:
	case OutFormText:
	case OutFormTable:
	case "":
		outFormat = OutFormJSON
	default:
		return nil, fmt.Errorf("provided output format isn't supported")
	}

	// Setting up request,
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	var cmds []interface{}
	cmds = append(cmds, *cmd)
	rpcReq := JSONRpcRequest{
		JSONRpcVersion: "2.0",
		ID:             id,
		Method:         MethodCli,
		Params: Params{
			Commands:  cmds,
			OutFormat: outFormat,
		},
	}
	// marshalling to []byte
	bRpcReq, err := json.Marshal(rpcReq)
	if err != nil {
		log.Fatal(err)
	}
	// ... creating an HTTP POST request
	reqHTTP, err := http.NewRequest("POST", fmt.Sprintf("https://%s:%v/jsonrpc", *t.Hostname, *t.PortJRpc), bytes.NewBuffer(bRpcReq))
	if err != nil {
		log.Fatal(err)
	}
	// setting content type and authentication header
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *t.Username, *t.Password))))

	tlsCfg := &tls.Config{InsecureSkipVerify: true} // Skipping verification
	// TODO: add certificate support

	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: tlsCfg,
	}}

	resp, err := client.Do(reqHTTP)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status: %s", resp.Status)
	}

	var rpcResp JSONRpcResponse
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, fmt.Errorf("decoding error: %s", err)
	}
	// Checking for RPC error presence
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("got an JSON-RPC error: %v", rpcResp.Error)
	}

	// Checking for id match
	if rpcResp.ID != id {
		return nil, fmt.Errorf("got an JSON-RPC error: %v", rpcResp.Error)
	}

	return &rpcResp, nil
}
