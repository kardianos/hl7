package hl7

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	v25 "github.com/kardianos/hl7/h250"
	v251 "github.com/kardianos/hl7/h251"
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

	e := NewEncoder(nil)
	bb, err := e.Encode(msg)
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

	d := NewDecoder(v25.Registry, nil)
	v, err := d.DecodeList(raw)
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

	d := NewDecoder(v25.Registry, nil)
	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	root, err := group(v, v25.Registry)
	if err != nil {
		t.Fatal(err)
	}
	e := NewEncoder(nil)
	rt, err := e.Encode(root)
	if err != nil {
		t.Fatal(err)
	}
	rt = bytes.ReplaceAll(rt, []byte{'\r'}, []byte{'\n'})
	if bytes.Equal(rt, raw) {
		t.Fatal("mismatch")
	}
}

func TestDecodeCompoundDateTime(t *testing.T) {
	flag.Parse()
	var raw = []byte(`MSH|^~\&|PATIENTPING_ADT|123456^Medical|1|uid-123456^^^PP^PP|20190306^^^default^default||JOE^DOE^||19541129|F|||31 MOZFA|272605|Medical|HOS|300 W 27th St^^Hometown^NC^28358|1790152668210992`)

	d := NewDecoder(v25.Registry, nil)
	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	printTypes(t, v)
	if len(v) != 1 {
		t.Fatal("expected one segment")
	}
	v0 := v[0]
	msg := v0.(*v25.MSH)
	t.Logf("MSH Date: %v", msg.DateTimeOfMessage)
}

func TestVaries(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "roundtrip", "vaers_long.hl7"))
	if err != nil {
		t.Fatal(err)
	}

	d := NewDecoder(v25.Registry, nil)
	v, err := d.Decode(raw)
	if err != nil {
		t.Fatal(err)
	}

	vg := v.(v25.ORU_R01)
	for _, pr := range vg.PatientResult {
		for _, oo := range pr.OrderObservation {
			for _, o := range oo.Observation {
				for _, v := range o.OBX.ObservationValue {
					t.Logf("Type: %s, Value: %#v", o.OBX.ValueType, v)
				}
			}
		}
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

	d := NewDecoder(v251.Registry, nil)
	e := NewEncoder(&EncodeOption{
		TrimTrailingSeparator: false,
	})

	for _, f := range dirList {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		t.Run(name, func(t *testing.T) {
			fn := filepath.Join(fsDir, name)
			bb, err := os.ReadFile(fn)
			if err != nil {
				t.Fatal(err)
			}
			v, err := d.DecodeList(bb)
			if err != nil {
				t.Fatal("unmarshal", err)
			}
			if *dumpSegmentList {
				c.Dump(v)
			}

			root, err := d.DecodeGroup(v)
			if err != nil {
				t.Fatal("group", err)
			}
			if *dumpSegmentGroup {
				c.Dump(root)
			}

			rt, err := e.Encode(root)
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

func TestError(t *testing.T) {
	fsDir := filepath.Join("testdata", "error")
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

	d := NewDecoder(v251.Registry, nil)

	for _, f := range dirList {
		if f.IsDir() {
			continue
		}
		const hl7Ext = ".hl7"
		name := f.Name()
		if !strings.HasSuffix(name, hl7Ext) {
			continue
		}
		errorName := name[:len(name)-len(hl7Ext)] + ".error"

		t.Run(name, func(t *testing.T) {
			fn := filepath.Join(fsDir, name)
			errorFn := filepath.Join(fsDir, errorName)

			bb, err := os.ReadFile(fn)
			if err != nil {
				t.Fatal(err)
			}
			errorBB, err := os.ReadFile(errorFn)
			if err != nil {
				t.Fatal(err)
			}

			var root any
			v, err := d.DecodeList(bb)
			if err == nil {
				if *dumpSegmentList {
					c.Dump(v)
				}

				root, err = d.DecodeGroup(v)
			}

			if err == nil && *dumpSegmentGroup {
				c.Dump(root)
			}

			if err == nil {
				t.Fatal("expected error, but got no error")
			}
			gotError := []byte(err.Error())

			if *overwrite {
				os.WriteFile(errorFn, gotError, 0600)
			}
			d := lineDiff(errorBB, gotError)
			if len(d) > 0 {
				t.Fatalf("mismatch\n%s", d)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	list := []struct {
		Name  string
		Input string
		Want  string
	}{
		{
			Name:  "normal1",
			Input: "20010330060500",
			Want:  "2001-03-30T06:05:00Z",
		},
		{
			Name:  "normal2",
			Input: "19991227140800",
			Want:  "1999-12-27T14:08:00Z",
		},
		{
			Name:  "ymd",
			Input: "19991227",
			Want:  "1999-12-27T00:00:00Z",
		},
		{
			Name:  "ymdhm",
			Input: "199912271408",
			Want:  "1999-12-27T14:08:00Z",
		},
		{
			Name:  "year",
			Input: "2001",
			Want:  "2001-01-01T00:00:00Z",
		},
		{
			Name:  "year-month",
			Input: "200110",
			Want:  "2001-10-01T00:00:00Z",
		},
		{
			Name:  "year-month-day",
			Input: "20011003",
			Want:  "2001-10-03T00:00:00Z",
		},
		// These are not properly formatted, but accept them anyway, the meaning is sufficently clear.
		{
			Name:  "bad-accept1",
			Input: "2019-07-02 12:23:24+0300",
			Want:  "2019-07-02T12:23:24+03:00",
		},
		{
			Name:  "bad-accept2",
			Input: "2019-07-02 12:23:24-0300",
			Want:  "2019-07-02T12:23:24-03:00",
		},
		// The meaning is not sufficently clear, reject.
		{
			Name:  "bad-length",
			Input: "2019-1",
			Want:  `unknown date time string size "20191"`,
		},
	}

	d := &lineDecoder{}
	for _, item := range list {
		t.Run(item.Name, func(t *testing.T) {
			var got string
			dt, err := d.parseDateTime(item.Input)
			if err != nil {
				got = err.Error()
			} else {
				got = dt.Format(time.RFC3339Nano)
			}
			if g, w := got, item.Want; g != w {
				t.Fatalf("got=%s want=%s", g, w)
			}
		})
	}
}
