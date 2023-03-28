package jsonrpc

import (
	"encoding/json"

	"github.com/azyablov/fat/lib/jsonrpc/datastores"
)

// class Command {
// 	<<interface>>
// 	+WithoutRecursion(): Command
// 	+WithDefaults(): Command
// 	+WithPathKeywords(jsonRawMessage): Command
// 	+WithDatastore(EnumDatastores): Command
// 	+GetDatastore(): EnumDatastores
// }

type Command interface {
	WithoutRecursion() Command
	WithDefaults() Command
	WithPathKeywords(json.RawMessage) Command
	WithDatastore(datastores.EnumDatastores) Command
	GetDatastore() Command
}

// note for GetCommand "Mandatory. List of commands used to execute against the called method. Multiple commands can be executed with a single request."
// class GetCommand {
// 	<<element>>
// 	note "Mandatory with the get, set and validate methods. This value is a string that follows the gNMI path specification1 in human-readable format."
// 	~string Path
// 	note "Optional, since can be embedded into path, for such kind of cases value should not be specified, so path assumed to follow <path>:<value> schema, which will be checked for set and validate"
// 	~string PathKeywords
// 	note "Optional; a Boolean used to retrieve children underneath the specific path. The default = true."
// 	~bool Recursive
// 	note "Optional; a Boolean used to show all fields, regardless if they have a directory configured or are operating at their default setting. The default = false."
// 	~bool Include-field-defaults
// 	+WithoutRecursion()
// 	+WithDefaults()
// 	+WithPathKeywords(jsonRawMessage) error
// 	+WithDatastore(EnumDatastores)
// }
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
	c.PathKeywords = jrm
	return nil
}

func (c *GetCommand) withDatastore(ds datastores.EnumDatastores) error {
	return c.Datastore(ds)
}
