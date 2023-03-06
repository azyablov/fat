package lib_test

import (
	"testing"

	"github.com/azyablov/fat/lib"
)

type CompElem interface {
	~int | float32 | float64 | string | byte
}

func TestGetSubStrPositions(t *testing.T) {
	var testString = `  test1 {
		"test2" {
			test3 {
                key value
				ip 8.8.8.8 
			}
		}
	}`
	pTest := lib.GetSubStrPositions(testString, "test")
	pKey := lib.GetSubStrPositions(testString, "key")
	pNl := lib.GetSubStrPositions(testString, "\n")
	pDik := lib.GetSubStrPositions(testString, "dik")

	if len(pTest) == 0 || len(pKey) == 0 || len(pNl) == 0 {
		t.Errorf("incorrect result: 0 positions found while substring is present")
	}
	if len(pDik) != 0 {
		t.Errorf("incorrect result: found string which is not part of the test sample")
	}
	if !(checkIntSlicesEqual(pTest, []int{2, 13, 25}) ||
		checkIntSlicesEqual(pKey, []int{49}) ||
		checkIntSlicesEqual(pNl, []int{9, 21, 32, 58, 74, 79, 83})) {
		t.Errorf("incorrect result: position returned aren't the expected ones")
	}
}

func checkIntSlicesEqual[T CompElem](x, y []T) bool {
	if len(x) != len(y) {
		return false
	}
	for i, v := range x {
		if v != y[i] {
			return false
		}
	}
	return true
}
