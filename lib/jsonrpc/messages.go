package jsonrpc

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/azyablov/fat/lib/jsonrpc/actions"
	"github.com/azyablov/fat/lib/jsonrpc/datastores"
	"github.com/azyablov/fat/lib/jsonrpc/formats"
	"github.com/azyablov/fat/lib/jsonrpc/methods"
)

// note for Command "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
//
//	class Command {
//		<<element>>
//		note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
//		~string Path
//		note "Optional, since can be embedded into path, for such kind of cases value should not be specified, so path assumed to follow <path>:<value> schema, which will be checked for set and validate"
//		~string Value
//		note "Optional; used to substitute named parameters with the path field. More than one keyword can be used with each path."
//		~string PathKeywords
//		note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
//		~bool Recursive
//		note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
//		~bool Include-field-defaults
//		+withoutRecursion()
//		+withDefaults()
//		+withPathKeywords(jsonRawMessage) error
//		+withDatastore(EnumDatastores)
//		+GetDatastore() string
//	}
//
// Command *-- "1" Action
// Command *-- "1" Datastore
type Command struct {
	Path                 string          `json:"path"`
	Value                string          `json:"value,omitempty"`
	PathKeywords         json.RawMessage `json:"path-keywords,omitempty"`
	Recursive            bool            `json:"recursive,omitempty"`
	IncludeFieldDefaults bool            `json:"include-field-defaults,omitempty"`
	*actions.Action
	*datastores.Datastore
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
	return c.Datastore.DSetDatastore(ds)
}

//	class CommandOption {
//		<<function>>
//		(Command c) error
//	}
type CommandOptions func(*Command) error

//	class CommandValue {
//		<<element>>
//		string
//	}
type CommandValue struct {
	string
}

// +NewCommand(EnumActions action, string path, string value, List~CommandOptions~ opts) sCommand
func NewCommand(action actions.EnumActions, path string, value CommandValue, opts ...CommandOptions) (*Command, error) {
	c := &Command{
		Path:      path,
		Recursive: true,
		Value:     value.string,
	}
	err := c.Action.SetAction(action)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
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
	*formats.OutputFormat
}

func (p *Params) appendCommands(commands []Command) {
	p.Commands = append(p.Commands, commands...)
}

func (p *Params) getCmds() *[]Command {
	return &p.Commands
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

func (r *Request) setID(id int) {
	r.ID = id
}

//	class Requester {
//		<<interface>>
//		+GetMethod() string
//		+Marshal() List~byte~
//		+GetID() int
//	}
type Requester interface {
	Marshal() ([]byte, error)
	GetMethod() (methods.EnumMethods, error)
	MethodName() string
	GetID() int
	appendCommands([]Command)
	setID(int)
}

//	class RequestOption {
//		<<function>>
//		(Request c) error
//	}
type RequestOption func(Requester) error

// +NewRequest(EnumMethods m, List~GetCommand~ cmds, List~RequestOption~ opts) Request
func NewRequest(m methods.EnumMethods, cmds []Command, opts ...RequestOption) (*Request, error) {
	r := &Request{}
	// set version
	r.JSONRpcVersion = "2.0"

	// set method
	err := r.Method.SetMethod(m)
	if err != nil {
		return nil, err
	}

	// set random ID
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)

	// set commands
	err = apply_cmds(r, cmds)
	if err != nil {
		return nil, err
	}

	// apply options to request
	err = apply_opts(r, opts)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func apply_cmds(r *Request, cmds []Command) error {
	// check if commands are empty
	if len(cmds) == 0 {
		return fmt.Errorf("no commands given")
	}

	// check if commands are valid for the selected method
	m, err := r.Method.GetMethod()
	if err != nil {
		return err
	}
	switch m {
	case methods.GET:
		for _, c := range cmds {
			if c.Action != nil {
				return fmt.Errorf("action not allowed for method %s", r.Method.MethodName())
			}
			if c.Value != "" {
				return fmt.Errorf("value not allowed for method %s", r.Method.MethodName())
			}
		}
		break
	case methods.SET || methods.VALIDATE:
		for _, c := range cmds {
			if c.Action == nil {
				return fmt.Errorf("action not found, but should be specified for method %s", r.Method.MethodName())
			}
			if c.Value == "" && !strings.Contains(c.Action, ":") {
				return fmt.Errorf("value isn't specified or not found in the path for method %s", r.Method.MethodName())
			}
		}
		break
	case methods.CLI:
		return fmt.Errorf("method %s not supported by Request, please use CLIRequest object", r.Method.MethodName())
	default:
		return fmt.Errorf("method %s not supported by Request", r.Method.MethodName())
	}
	// checks passed, append commands to request
	r.appendCommands(cmds)
	return nil
}

// function applies options to the request
func apply_opts(r Requester, opts []RequestOption) error {
	for _, o := range opts {
		if err := o(r); err != nil {
			return nil
		}
	}
	return nil
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
	*formats.OutputFormat
}

func (p *CLIParams) appendCommands(commands []string) {
	p.Commands = append(p.Commands, commands...)
}

func (p *CLIParams) getCmds() *[]string {
	return &p.Commands
}

// note for CLIRequest "JSON RPC Request: cli"
//
//	class CLIRequest {
//		<<message>>
//		note "Method set to CLI"
//		note "Mandatory. Version, which must be ‟2.0”. No other JSON RPC versions are currently supported."
//		~string JSONRpcVersion
//		note "Mandatory. Client-provided integer. The JSON RPC responds with the same ID, which allows the client to match requests to responses when there are concurrent requests."
//		~int ID
//		note "Mandatory. Supported options are cli. Set statically in the RPC request"
//		+Marshal() List~byte~
//		+GetID() int
//		~setID(int)
//	}
//
// CLIRequest *-- Method
// CLIRequest *-- CLIParams
type CLIRequest struct {
	JSONRpcVersion string `json:"jsonrpc"`
	ID             int    `json:"id"`
	methods.Method
	CLIParams
}

func (r *CLIRequest) Marshal() ([]byte, error) {
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (r *CLIRequest) GetID() int {
	return r.ID
}

func (r *CLIRequest) setID(id int) {
	r.ID = id
}

// +NewCLIRequest(List~string~ cmds, List~RequestOption~ opts) CLIRequest
func NewCLIRequest(cmds []string, opts ...RequestOption) (*CLIRequest, error) {
	r := &CLIRequest{}
	// set version
	r.JSONRpcVersion = "2.0"

	// set method
	err := r.Method.SetMethod(methods.CLI)
	if err != nil {
		return nil, err
	}

	// set random ID
	rand.Seed(time.Now().UnixNano())
	id := rand.Int()
	r.setID(id)

	// set commands
	r.appendCommands(cmds)

	// apply options to request
	err = apply_opts(r, opts)
	if err != nil {
		return nil, err
	}

	return r, nil
}