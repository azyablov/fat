package lib

import (
	"strings"
	"time"
)

type SSHAttr struct {
	NoStrictKey *bool
}

type TLSAttr struct {
	RootCA     *string // CA certificate file in PEM format.
	Cert       *string // Client certificate file in PEM format.
	Key        *string // Client private key file.
	InsecConn  *bool   // Insecure connection.
	SkipVerify *bool   // Diable certificate validation during TLS session ramp-up.
}

type Cred struct {
	Username *string
	Password *string
}

type TargetHost struct {
	Hostname *string
	PortSSH  *int
	PortgNOI *int
	PortgNMI *int
	PortJRpc *int
	Timeout  *time.Duration
}

type SRLTarget struct {
	TargetHost
	Cred
	SSHAttr
	TLSAttr
}

// Function returns the list of indexes substring within provided string or empty slice, if no substring found.
func GetSubStrPositions(str string, sub string) []int {
	var subInd []int = make([]int, 0, 64)
	var s, i int
	var l = len(str)

	for i = strings.Index(str[s:], sub); i != -1; {
		subInd = append(subInd, s+i)
		// Shift to the position after \n
		s = s + i + 1
		if s >= l {
			return subInd
		}
		i = strings.Index(str[s:], sub)
	}
	return subInd
}
