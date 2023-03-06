package lib

import "strings"

type SSHAttr struct {
	NoStrictKey *bool
}
type Cred struct {
	Username *string
	Password *string
}

type TargetHost struct {
	Hostname *string
	Port     *int
}

type SRLTarget struct {
	TargetHost
	Cred
	SSHAttr
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
