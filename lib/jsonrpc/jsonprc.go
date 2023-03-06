package jsonrpc

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

const (
	MethodGet Method = "get"
	MethodSet Method = "set"
	MethodCli Method = "cli"
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
	Commands     []interface{} `json:"commands"`
	OutputFormat string        `json:"output-format,omitempty"`
}

type RpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func ExecCli(t *lib.SRLTarget, cmd string) (*JSONRpcResponse, error) {

	if len(cmd) == 0 {
		return nil, fmt.Errorf("command can't be null string")
	}

	// Setting up request,
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	var cmds []interface{}
	cmds[0] = cmd
	rpcReq := JSONRpcRequest{
		JSONRpcVersion: "2.0",
		ID:             id,
		Method:         MethodCli,
		Params: Params{
			Commands: cmds,
		},
	}
	// marshalling to []byte
	bRpcReq, err := json.Marshal(rpcReq)
	if err != nil {
		log.Fatal(err)
	}
	// ... creating an HTTP POST request
	reqHTTP, err := http.NewRequest("POST", fmt.Sprintf("%s:%s", *t.Hostname, *t.Port), bytes.NewBuffer(bRpcReq))
	if err != nil {
		log.Fatal(err)
	}
	// setting content type and authentication header
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", *t.Username, *t.Password))))

	tlsCfg := &tls.Config{InsecureSkipVerify: false} // Skipping verification
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
		log.Fatal("Status: %s", resp.Status)
	}

	var rpcResp JSONRpcResponse
	err = json.NewDecoder(resp.Body).Decode(rpcResp)
	if err != nil {
		log.Fatal(err)
	}

	if rpcResp.Error != nil {
		log.Fatalf("JSON-RPC error:", rpcResp.Error)
	}

	return &rpcResp, nil
}

// {
// 	"jsonrpc": "2.0",
// 	"id": 0,
// 	"method": "cli",
// 	"params": {
// 	  "commands": [
// 		"enter candidate",
// 		"info interface mgmt0"
// 	  ]
// 	}
//   }
