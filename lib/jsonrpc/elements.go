package jsonrpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/azyablov/fat/lib/jsonrpc/datastores"
	"github.com/azyablov/fat/lib/jsonrpc/formats"
	"github.com/azyablov/fat/lib/jsonrpc/methods"
	"github.com/openconfig/gnmic/actions"
)

// note for Command "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
//
//	class Command {
//		<<element>>
//		note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
//		~string Path
//		note "Optional; used to substitute named parameters with the path field. More than one keyword can be used with each path."
//		~string PathKeywords
//		note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
//		~bool Recursive
//		note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
//		~bool Include-field-defaults
//		+withoutRecursion(): Command
//		+withDefaults(): Command
//		+withPathKeywords(jsonRawMessage): Command
//		+withDatastore(EnumDatastores): Command
//	}
//
// Command *-- "1" Datastore
type Command struct {
	Path                 string          `json:"path"`
	Value                string          `json:"value,omitempty"`
	PathKeywords         json.RawMessage `json:"path-keywords,omitempty"`
	Recursive            bool            `json:"recursive,omitempty"`
	IncludeFieldDefaults bool            `json:"include-field-defaults,omitempty"`
	actions.Action
	datastores.Datastore
}

func (c *Command) withoutRecursion() {
	c.Recursive = false
}

func (c *Command) withDefaults() {
	c.IncludeFieldDefaults = true
}

func (c *Command) withPathKeywords(jrm json.RawMessage) error {
	var data interface{}
	err := json.Unmarshal(jrm, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal path-keywords: %v", err)
	}
	c.PathKeywords = jrm
	return nil
}

func (c *Command) withDatastore(ds datastores.EnumDatastores) error {
	return c.SetDatastore(ds)
}

// note for params "MAY be omitted. Defines a container for any parameters related to the request. The type of parameter is dependent on the method used."
//
//	class Params {
//		<<element>>
//		~List~Command~ commands
//		+appendCommands(List~Command~)
//	}
//
// Params *-- OutputFormat
type Params struct {
	Commands []Command `json:"commands"`
	formats.OutputFormat
}

func (p *Params) appendCommands(commands []Command) {
	p.Commands = append(p.Commands, commands...)
}

//	class CLIParams {
//		<<element>>
//		~List~string~ commands
//		+appendCommands(List~string~)
//	}
//
// CLIParams *-- OutputFormat
type CLIParams struct {
	Commands []string `json:"commands"`
	formats.OutputFormat
}

func (p *CLIParams) appendCommands(commands []string) {
	p.Commands = append(p.Commands, commands...)
}

// note for RpcError "When a rpc call is made, the Server MUST reply with a Response, except for in the case of Notifications. The Response is expressed as a single JSON Object"
//
//	class RpcError {
//		<<element>>
//		note "A Number that indicates the error type that occurred. This MUST be an integer."
//		+int ID
//		note "A String providing a short description of the error. The message SHOULD be limited to a concise single sentence."
//		+string Message
//		note "A Primitive or Structured value that contains additional information about the error. This may be omitted. The value of this member is defined by the Server (e.g. detailed error information, nested errors etc.)."
//		+string Data
//	}
type RpcError struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

//	class Requester {
//		<<interface>>
//		+GetMethod() string
//		+Marshal() List~byte~
//		+GetID() int
//	}
type Requester interface {
	GetMethod() string
	Marshal() ([]byte, error)
	GetID() int
	setID(int)
	appendCommands([]Command)
	//TODO: extend interface for options!!!
}

// note for Request "JSON RPC Request: get / set / validate"
//
//	class Request {
//		<<message>>
//		note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
//		~string JSONRpcVersion
//		note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
//		~int ID
//		+Marshal() List~byte~
//		+GetID() int
//	}
//
// Request *-- Method
// Request *-- Params
type Request struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	methods.Method
	Params
	//TODO: extend struct for options!!!
}

func (r *Request) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *Request) GetID() int {
	return r.ID
}

func (r *Request) GetMethod() string {
	return r.Method.GetMethod()
}

func (r *Request) setID(id int) {
	r.ID = id
}

//	class RequestOption {
//		<<function>>
//		(Request c) error
//	}
type RequestOption func(Requester) error

//	class GetRequest {
//		<<message>>
//		note "Method set to GET"
//	}
//
// GetRequest *-- "1" Request
type GetRequest struct {
	Request
}

// +NewGetRequest(List~GetCommand~ cmds, List~RequestOption~ opts) GetRequest
func NewGetRequest(cmds []Command, opts ...RequestOption) (*GetRequest, error) {
	r := &GetRequest{}
	r.Method.SetMethod(methods.GET)
	err := apply_opts_and_cmds(r, cmds, opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//	class SetRequest {
//		<<message>>
//		note "Method set to SET"
//	}
//
// SetRequest *-- "1" Request
type SetRequest struct {
	Request
}

// +NewSetRequest(List~SetCommand~ cmds, List~RequestOption~ opts) SetRequest
func NewSetRequest(cmds []Command, opts ...RequestOption) (*SetRequest, error) {
	r := &SetRequest{}
	r.Method.SetMethod(methods.SET)

	err := apply_opts_and_cmds(r, cmds, opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}

//	class ValidateRequest {
//		<<message>>
//		note "Method set to VALIDATE"
//	}
//
// ValidateRequest *-- "1" Request
type ValidateRequest struct {
	Request
}

// +NewValidateRequest(List~ValidateCommand~ cmds, List~RequestOption~ opts) ValidateRequest
func NewValidateRequest(cmds []Command, opts ...RequestOption) (*ValidateRequest, error) {
	r := &ValidateRequest{}
	r.Method.SetMethod(methods.VALIDATE)

	err := apply_opts_and_cmds(r, cmds, opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func apply_opts_and_cmds(r Requester, cmds []Command, opts []RequestOption) error {
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)
	r.appendCommands(cmds)
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil
		}
	}
	return nil
}
