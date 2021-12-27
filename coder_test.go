package hl7

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	v25 "github.com/kardianos/hl7/v25"
	v251 "github.com/kardianos/hl7/v251"
	"github.com/sanity-io/litter"
)

var overwrite = flag.Bool("overwrite", false, "overwrite testing values with got values")
var dumpSegmentList = flag.Bool("dump-seg", false, "show the segment list dump")
var dumpSegmentGroup = flag.Bool("dump-group", false, "show the segment group dump")

func TestEncode(t *testing.T) {
	msg := v25.MSH{
		FieldSeparator:     `|`,
		EncodingCharacters: `^~\&`,
		VersionID: v25.VID{
			VersionID: "2.5",
		},
	}

	bb, err := Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	bb = bytes.ReplaceAll(bb, []byte{'\r'}, []byte{'\n'})
	t.Log(string(bb))
}

func TestDecode(t *testing.T) {
	var raw = []byte(`MSH|^~\&||||||||||2.5^^|||||||||
PID|1||||^Bob||||||||||||||||||||||||||||||||||
NTE|1||testing the system comments here|||
NTE|2||more comments here|||
OBR|1|ABC||||||||||||||1234^Acron^Smith~5678^Beta^Zeta|||||||||||||||||||||||||||||||||||||||||||
OBR|2|XYZ||||||||||||||903^Blacky|||||||||||||||||||||||||||||||||||||||||||`)

	v, err := Unmarshal(raw, UnmarshalOption{
		Registry: v25.Registry,
	})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	_ = data
}
func TestGroup(t *testing.T) {
	flag.Parse()
	var raw = []byte(`MSH|^~\&|DX_LAB|Hematology|WPX||20070305170957||ORL^O34^ORL_O34|2|P|2.5||||||8859/1|||
MSA|AA|161||||
`)

	v, err := Unmarshal(raw, UnmarshalOption{
		Registry: v25.Registry,
	})
	if err != nil {
		t.Fatal(err)
	}
	root, err := Group(v, v25.Registry)
	if err != nil {
		t.Fatal(err)
	}
	rt, err := Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	rt = bytes.ReplaceAll(rt, []byte{'\r'}, []byte{'\n'})
	if bytes.Equal(rt, raw) {
		t.Fatal("mismatch")
	}
}

func printTypes(t *testing.T, list []any) {
	for _, item := range list {
		t.Logf("type %T", item)
	}
}

func (w *walker) printLinear(t *testing.T) {
	for index, item := range w.list {
		t.Logf("linear[%d] %s leaf=%t inArray=%t", index, item.Type.String(), item.Leaf, item.InArray)
	}
}

func TestRoundTrip(t *testing.T) {
	fsDir := filepath.Join("testdata", "roundtrip")
	dirList, err := os.ReadDir(fsDir)
	if err != nil {
		t.Fatal(err)
	}
	c := &litter.Options{
		HideZeroValues:    true,
		HidePrivateFields: true,
		Separator:         " ",
	}
	_ = c
	for _, f := range dirList {
		if f.IsDir() {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			fn := filepath.Join(fsDir, f.Name())
			bb, err := os.ReadFile(fn)
			if err != nil {
				t.Fatal(err)
			}
			uo := UnmarshalOption{
				Registry: v251.Registry,
			}
			v, err := Unmarshal(bb, uo)
			if err != nil {
				t.Fatal("unmarshal", err)
			}
			if *dumpSegmentList {
				c.Dump(v)
			}

			root, err := Group(v, v251.Registry)
			if err != nil {
				t.Fatal("group", err)
			}
			if *dumpSegmentGroup {
				c.Dump(root)
			}

			rt, err := Marshal(root)
			if err != nil {
				t.Fatal("marshal", err)
			}
			rt = bytes.ReplaceAll(rt, []byte{'\r'}, []byte{'\n'})
			if *overwrite {
				os.WriteFile(fn, rt, 0600)
			}
			d := lineDiff(bb, rt)
			if len(d) > 0 {
				t.Fatalf("mismatch\n%s", d)
			}
		})
	}
}
