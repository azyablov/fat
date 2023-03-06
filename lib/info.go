package lib

import (
	"fmt"
	"strings"
)

type InfoObject struct {
	Key     string
	StLine  int
	EndLine int
	StInd   int
	EndInd  int
	Chlds   []*InfoObject
}

// Creates new InfoObject from the provided string.
func NewInfoObject(info string) (*InfoObject, error) {

	nLines := GetSubStrPositions(info, "\n")
	if len(nLines) == 0 {
		return nil, fmt.Errorf("malformed info; no blocks found")
	}
	// Creating virtual root.
	rootedInfo := strings.Join([]string{"root {", info, "}\n"}, "\n")

	p, err := parseToInfoObjTree(rootedInfo, GetSubStrPositions(rootedInfo, "\n"), 0)
	if err != nil {
		return nil, err
	}
	return p, nil

}

// Creates new InfoObject from the provided string w/o virtual root.
func NewInfoObjectWOvRoot(info string) (*InfoObject, error) {

	nLines := GetSubStrPositions(info, "\n")
	if len(nLines) == 0 {
		return nil, fmt.Errorf("malformed info; no blocks found")
	}

	p, err := parseToInfoObjTree(info, nLines, 0)
	if err != nil {
		return nil, err
	}
	return p, nil

}

// Function parses provided SR Linux and returns InfoObject tree
func parseToInfoObjTree(s string, nLines []int, start int) (*InfoObject, error) {
	// A new block object.
	var block InfoObject
	// Starting pointer of the string to find index.
	sol := 0
	// End of child block to seek starting pointer.
	chldEnd := 0
	// Start of the block found FLAG.
	var bs bool
	for line, eol := range nLines[start:] {
		if start+line != 0 {
			sol = nLines[start+line-1] + 1
		}
		// skip chld blocks, if found
		if start+line < chldEnd {
			continue
		}
		if bs {
			// block start identified
			crlBrSt := strings.Index(s[sol:eol], "{")
			crlBrEnd := strings.Index(s[sol:eol], "}")
			switch {
			case crlBrSt == -1 && crlBrEnd == -1:
				// Config element, skip and move to the next line.
				// Shifting start of the line to the position right after eol.
				continue
			case crlBrSt != -1 && crlBrEnd != -1:
				// start and end of the block on the same line
				return nil, fmt.Errorf("malformed info; start and end of the block on same line %+v", start+line)
			case crlBrSt != -1 && crlBrEnd == -1:
				// Found start of the new block.
				chld, err := parseToInfoObjTree(s, nLines, start+line)
				if err != nil {
					return nil, err
				}
				block.Chlds = append(block.Chlds, chld)
				chldEnd = chld.EndLine + 1
				continue
			case crlBrSt == -1 && crlBrEnd != -1:
				// found block end
				block.EndLine = start + line
				block.EndInd = eol
				if block.Key == "root" && line != len(nLines)-1 {
					// Conditions means we have reached end of root block, but didn't reach end of the info config
					return nil, fmt.Errorf("malformed info; { } aren't matching each other correctly at end of info config")
				}
				return &block, nil
			default:
				return nil, fmt.Errorf("malformed info; unexpected error on the line %+v", start+line)
			}

		} else {
			// Looking for the block start / end.
			crlBrSt := strings.Index(s[sol:eol], "{")
			crlBrEnd := strings.Index(s[sol:eol], "}")
			// No start of the block found.
			if crlBrSt == -1 {
				// In case found block end w/o block start, panic.
				if crlBrEnd != -1 {
					return nil, fmt.Errorf("malformed info; found block end w/o block start, check for virtual root")
				}
				return nil, fmt.Errorf("malformed info; supposed to see start of the block, but not found, check for virtual root")
			} else {
				// found block start, setting flag and record line #.
				if crlBrEnd != -1 {
					return nil, fmt.Errorf("malformed info; start and end of the block on the same line")
				}
				block.StLine = start + line
				block.StInd = sol
				// fmt.Println("Start of the block: ", s[sol:eol])
				block.Key = strings.TrimSpace(s[sol : sol+crlBrSt])
				bs = true
			}
			// Shifting start of the line to the position right after eol.
			continue
		}
	}
	return &block, fmt.Errorf("malformed info; missed end of the block OR unexpected error")
}

// Function is removing clab related config from the info tree, except /interface, /system/aaa /system/lldp parts which are usually a part of lab modelling
// set / system tls server-profile clab-profile key "{{ .TLSKey }}"
// set / system tls server-profile clab-profile certificate "{{ .TLSCert }}"
// {{- if .TLSAnchor }}
// set / system tls server-profile clab-profile authenticate-client true
// set / system tls server-profile clab-profile trust-anchor "{{ .TLSAnchor }}"
// {{- else }}
// set / system tls server-profile clab-profile authenticate-client false
// {{- end }}
// set / system gnmi-server admin-state enable network-instance mgmt admin-state enable tls-profile clab-profile
// set / system gnmi-server rate-limit 65000
// set / system gnmi-server trace-options [ request response common ]
// set / system gnmi-server unix-socket admin-state enable
// set / system json-rpc-server admin-state enable network-instance mgmt http admin-state enable
// set / system json-rpc-server admin-state enable network-instance mgmt https admin-state enable tls-profile clab-profile
// set / system lldp admin-state enable
// set / system aaa authentication idle-timeout 7200
// {{/* enabling interfaces referenced as endpoints for a node (both e1-2 and e1-3-1 notations) */}}
// {{- range $ep := .Endpoints }}
// {{- $parts := ($ep.EndpointName | strings.ReplaceAll "e" "" | strings.Split "-") -}}
// set / interface ethernet-{{index $parts 0}}/{{index $parts 1}} admin-state enable
//
//	{{- if eq (len $parts) 3 }}
//
// set / interface ethernet-{{index $parts 0}}/{{index $parts 1}} breakout-mode num-channels 4 channel-speed 25G
// set / interface ethernet-{{index $parts 0}}/{{index $parts 1}}/{{index $parts 2}} admin-state enable
//
//	{{- end }}
//
// {{ end -}}
// set / system banner login-banner "{{ .Banner }}"
func CleanUpClabInfoObjects(root *InfoObject, s string) (string, error) {
	var sliceInx = make([]int, 1, 10)
	var system *InfoObject
	var sanStr []string

	// Protection from being provided with empty root.
	if len(root.Chlds) == 0 {
		return "", fmt.Errorf("no child objects under virtual root, nothing to do")
	}

	// Protection from being provided non root.
	if root.Key != "root" {
		return "", fmt.Errorf("root InfoObject should be with virtual root")
	}

	for _, c := range root.Chlds {
		if c.Key == "system" {
			system = c
			break
		}
	}
	if system == nil {
		return "", fmt.Errorf("unable to find system elem in tree")
	}

	for _, c := range system.Chlds {
		switch c.Key {
		case "tls":
			for _, c := range c.Chlds {
				if c.Key == "server-profile clab-profile" {
					sliceInx = append(sliceInx, c.StInd, c.EndInd)
				}
			}
		case "gnmi-server", "json-rpc-server", "banner":
			sliceInx = append(sliceInx, c.StInd, c.EndInd)
		}
	}
	sliceInx = append(sliceInx, len(s))

	for i := 0; i < len(sliceInx); i += 2 {
		sanStr = append(sanStr, s[sliceInx[i]:sliceInx[i+1]])
	}

	return strings.Join(sanStr, ""), nil
}

// Func prints object tree to the terminal.
func PrintInfObjTree(i *InfoObject, ident ...int) {
	if len(ident) == 0 {
		ident = append(ident, 2)
	}
	if i.Key != "root" {
		fmt.Print(strings.Repeat(" ", ident[0]), "└─", i.Key, "\n")
	} else {
		fmt.Print(strings.Repeat(" ", ident[0]), i.Key, "\n")
		ident = append(ident, ident[0])
	}
	for n := range i.Chlds {
		PrintInfObjTree(i.Chlds[n], ident[0]+ident[1], ident[1])
	}
}
