package jsonrpc

import (
	"encoding/json"
	"fmt"

	"github.com/azyablov/fat/lib/jsonrpc/datastores"
	"github.com/azyablov/fat/lib/jsonrpc/formats"
)

//	class Command {
//		<<interface>>
//		+WithoutRecursion(): Command
//		+WithDefaults(): Command
//		+WithPathKeywords(jsonRawMessage): Command
//		+WithDatastore(EnumDatastores): Command
//		+GetDatastore(): EnumDatastores
//	}
type Command interface {
	withoutRecursion() Command
	withDefaults() Command
	withPathKeywords(json.RawMessage) Command
	withDatastore(datastores.EnumDatastores) Command
	GetDatastore() Command
}

// note for GetCommand "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
//
//	class GetCommand {
//		<<element>>
//		note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
//		~string Path
//		note "Optional, since can be embedded into path, for such kind of cases value should not be specified, so path assumed to follow <path>:<value> schema, which will be checked for set and validate"
//		~string PathKeywords
//		note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
//		~bool Recursive
//		note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
//		~bool Include-field-defaults
//		+WithoutRecursion()
//		+WithDefaults()
//		+WithPathKeywords(jsonRawMessage) error
//		+WithDatastore(EnumDatastores)
//	}
//
// GetCommand *-- "1" Datastore
type GetCommand struct {
	Path                 string          `json:"path"`
	PathKeywords         json.RawMessage `json:"path-keywords,omitempty"`
	Recursive            bool            `json:"recursive,omitempty"`
	IncludeFieldDefaults bool            `json:"include-field-defaults,omitempty"`
	datastores.Datastore
}

func (c *GetCommand) withoutRecursion() {
	c.Recursive = false
}

func (c *GetCommand) withDefaults() {
	c.IncludeFieldDefaults = true
}

func (c *GetCommand) withPathKeywords(jrm json.RawMessage) error {
	var data interface{}
	err := json.Unmarshal(jrm, &data)
	if err != nil {
		return fmt.Errorf("failed to unmarshal path-keywords: %v", err)
	}
	c.PathKeywords = jrm
	return nil
}

func (c *GetCommand) withDatastore(ds datastores.EnumDatastores) error {
	return c.SetDatastore(ds)
}

// note for params "MAY be omitted. Defines a container for any parameters related to the request. The type of parameter is dependent on the method used."
//
//	class Params {
//		<<element>>
//		~List~Command~ commands
//	}
//
// Params *-- OutputFormat
type Params struct {
	Commands []Command `json:"commands"`
	formats.OutputFormat
}

//	class CLIParams {
//		<<element>>
//		~List~string~ commands
//	}
//
// CLIParams *-- OutputFormat
type CLIParams struct {
	Commands []string `json:"commands"`
	formats.OutputFormat
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
