package lib_test

import (
	"os"
	"strings"
	"testing"

	"github.com/azyablov/fat/lib"
	"github.com/google/go-cmp/cmp"
)

const (
	sampleInfoObjFile          = "./testdata/infosample.cfg"
	sampleNoNewLines           = "./testdata/nonewlines.cfg"
	sampleNoBlocks             = "./testdata/noblocks.cfg"
	sampleEndWOStart           = "./testdata/endwostart.cfg"
	sampleStartWOEnd           = "./testdata/startwoend.cfg"
	sampleNoBlocksML           = "./testdata/noblockmultiline.cfg"
	sampleEndWOStartInTheMid   = "./testdata/endwostartinthemid.cfg"
	sampleMissedEndOfTheBlock  = "./testdata/missedendofblock.cfg"
	sampleEndOfTheBLockWOStart = "./testdata/endoftheblockwostart.cfg"
	sampleNoSystem             = "./testdata/nosystem.cfg"
	sampleSystem               = "./testdata/system.cfg"
)

func TestNewInfoObject(t *testing.T) {
	// Expected results
	var chldLevel1 = []string{"interface system0", "system", "network-instance MAC-VRF-3"}
	var chldLevel2 = []string{"subinterface 0", "aaa", "lldp", "gnmi-server", "tls", "json-rpc-server", "clock", "ssh-server", "banner", "logging", "network-instance", "interface lag1.0", "vxlan-interface vxlan3.1", "protocols"}
	var chldLevel7 = []string{"ethernet-segment client2", "mac-ip", "inclusive-mcast"}
	bs, err := os.ReadFile(sampleInfoObjFile)
	if err != nil {
		t.Fatalf("can't read test data: %+v", err)
	}

	infoObj, err := lib.NewInfoObject(string(bs))
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(chldLevel1, getKeysOfInfObjLvl(infoObj, 1)); diff != "" {
		t.Errorf("NewInfoObject() mismatch (-chldLevel1 +getKeysOfInfObjLvl()):\n%s", diff)
	}

	if diff := cmp.Diff(chldLevel2, getKeysOfInfObjLvl(infoObj, 2)); diff != "" {
		t.Errorf("NewInfoObject() mismatch (-chldLevel2 +getKeysOfInfObjLvl()):\n%s", diff)
	}

	if diff := cmp.Diff(chldLevel7, getKeysOfInfObjLvl(infoObj, 7)); diff != "" {
		t.Errorf("NewInfoObject() mismatch (-chldLevel7 +getKeysOfInfObjLvl()):\n%s", diff)
	}

	testData := []struct {
		testName string
		file     string
		expErr   string
	}{
		{testName: "Checking err: start and end of the block of the same line", file: sampleNoNewLines, expErr: "start and end of the block on same line"},
		{testName: "No blocks error SL", file: sampleNoBlocks, expErr: "no blocks found"},
		{testName: "Checking err: } w/o { in the middle of info config", file: sampleEndWOStartInTheMid, expErr: "{ } aren't matching each other correctly at end of info config"},
		{testName: "Checking err: } w/o { at end of info config", file: sampleEndWOStart, expErr: "{ } aren't matching each other correctly at end of info config"},
		{testName: "Checking err: { w/o }", file: sampleStartWOEnd, expErr: "missed end of the block OR unexpected error"},
		{testName: "Checking err: no start of the block {", file: sampleNoBlocksML, expErr: "supposed to see start of the block, but not found, check for virtual root"},
		{testName: "Checking err: missed end of the block", file: sampleMissedEndOfTheBlock, expErr: "missed end of the block OR unexpected error"},
		{testName: "Checking err: end of the block w/o start {", file: sampleEndOfTheBLockWOStart, expErr: "found block end w/o block start, check for virtual root"},
	}
	for n, d := range testData {
		t.Run(d.testName, func(t *testing.T) {
			expErr := d.expErr
			bs, err = os.ReadFile(d.file)
			if err != nil {
				t.Fatalf("can't read test data: %+v", err)
			}

			_, err = lib.NewInfoObject(string(bs))
			if n >= 5 {
				_, err = lib.NewInfoObjectWOvRoot(string(bs))
			}
			t.Log(err)
			if err != nil {
				if !strings.Contains(err.Error(), expErr) {
					t.Errorf("expected error: %s; got: %v\n", expErr, err)
				}
			} else {
				t.Errorf("incorrect error handling for: %s; got: %v\n", expErr, err)
			}
		})
	}

}

func TestCleanUpClabInfoObjects(t *testing.T) {
	bs, err := os.ReadFile(sampleInfoObjFile)
	if err != nil {
		t.Fatalf("can't read test data: %+v", err)
	}

	infoObj, err := lib.NewInfoObject(string(bs))
	if err != nil {
		t.Fatalf("got an error from NewInfoObject(): %+v\n", err)
	}
	actStr, err := lib.CleanUpClabInfoObjects(infoObj, string(bs))
	if err != nil {
		t.Fatalf("got an error from CleanUpClabInfoObjects(): %+v\n", err)
	}
	switch {
	case strings.Contains(actStr, `server-profile clab-profile`):
		t.Errorf("incorrect result: tls profile is part of the config")
	case strings.Contains(actStr, `tls-profile clab-profile`):
		t.Errorf("incorrect result: clab-profile block is not removed the config")
	case strings.Contains(actStr, `gnmi-server`):
		t.Errorf("incorrect result: gnmi-server block is part of the config")
	case strings.Contains(actStr, `certificate "`):
		t.Errorf("incorrect result: certificate is not removed the config")
	case strings.Contains(actStr, `json-rpc-server`):
		t.Errorf("incorrect result: json-rpc-server block is not removed the config")
	}

	testData := []struct {
		testName string
		file     string
		expErr   string
	}{
		{testName: "Checking err: no system element", file: sampleNoSystem, expErr: "unable to find system elem in tree"},
		{testName: "Checking err: no virtual root", file: sampleSystem, expErr: "root InfoObject should be with virtual root"},
	}
	for n, d := range testData {
		t.Run(d.testName, func(t *testing.T) {
			expErr := d.expErr
			bs, err = os.ReadFile(d.file)
			if err != nil {
				t.Fatalf("can't read test data: %+v", err)
			}

			infoObj, err = lib.NewInfoObject(string(bs))
			if n >= 1 {
				infoObj, err = lib.NewInfoObjectWOvRoot(string(bs))
			}
			_, err := lib.CleanUpClabInfoObjects(infoObj, string(bs))
			t.Log(err)
			if err != nil {
				if !strings.Contains(err.Error(), expErr) {
					t.Errorf("expected error: %s; got: %v\n", expErr, err)
				}
			} else {
				t.Errorf("incorrect error handling for: %s; got: %v\n", expErr, err)
			}
		})
	}

}

func getKeysOfInfObjLvl(i *lib.InfoObject, lvl int) []string {
	var keys []string
	if lvl < 1 {
		return keys
	}

	var pBlocks, cBlocks []*lib.InfoObject
	pBlocks = append(pBlocks, i)

	for l := 1; lvl >= l; l++ {
		for _, b := range pBlocks {
			cBlocks = append(cBlocks, b.Chlds...)
		}
		pBlocks = cBlocks
		cBlocks = []*lib.InfoObject{}
	}
	for _, b := range pBlocks {
		keys = append(keys, b.Key)
	}
	return keys
}
