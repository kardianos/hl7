package hl7

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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

// DecodeInlineError test that a single field error will not prevent
// decoding the remainder of the message.
func TestDecodeInlineError(t *testing.T) {
	// The date of birth is incorrect in this PID. The day of month is "92" rather then "02".
	// 92 is an invalid day of month, so it will error. However, only this segment will error, and the other valid segment fields
	// will still be accessable.
	var raw = []byte(`MSH|^~\&|DX_LAB|Hematology|WPX||20070305170957||ORL^O34^ORL_O34|2|P|2.5||||||8859/1|||
PID|1||PID1992299||Smith^John||19561192000000|M||Caucasian||||||||||||Caucasian|||||||||||||||||
NTE|1||testing the system comments here|||
NTE|2||more comments here|||
OBR|1|ABC||||||||||||||1234^Acron^Smith~5678^Beta^Zeta|||||||||||||||||||||||||||||||||||||||||||
OBR|2|XYZ||||||||||||||903^Blacky|||||||||||||||||||||||||||||||||||||||||||`)

	d := NewDecoder(v25.Registry, nil)
	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	const (
		wantFirst = "John"
		wantLast  = "Smith"
		wantError = `line 2, PID.DateTimeOfBirth(time.Time)[7]: parsing time "19561192000000": day out of range`
	)
	var gotError, gotFirst, gotLast string
	for _, item := range v {
		if se, ok := item.(SegmentError); ok {
			gotError = errors.Join(se.ErrorList...).Error()
			if pid, ok := se.Segment.(*v25.PID); ok {
				if len(pid.PatientName) > 0 {
					p := pid.PatientName[0]
					gotFirst = p.GivenName
					gotLast = p.FamilyName
				}
			}
		}
	}
	ck := func(name, g, w string) {
		if g != w {
			t.Errorf("for %s, got=<<%s>> want=<<%s>>", name, g, w)
		}
	}
	ck("error", gotError, wantError)
	ck("first", gotFirst, wantFirst)
	ck("last", gotLast, wantLast)

	gr, err := d.DecodeGroup(v)
	var grErrText string
	if err != nil {
		grErrText = err.Error()
	}
	ck("group-error", grErrText, wantError)
	m, ok := gr.(v25.ORL_O34)
	if !ok {
		t.Fatal("incorrect message type")
	}

	p := m.Response.Patient.PID.PatientName[0]
	gotFirst = p.GivenName
	gotLast = p.FamilyName
	ck("first", gotFirst, wantFirst)
	ck("last", gotLast, wantLast)
}

// Test decoding only the header.
func TestDecodeHeader(t *testing.T) {
	var raw = []byte(`MSH|^~\&|DX_LAB|Hematology|WPX||20070305170957|XYZ|ORL^O34^ORL_O34|2|P|2.5||||||8859/1|||
PID|1||PID1992299||Smith^John||19561192000000|M||Caucasian||||||||||||Caucasian|||||||||||||||||
NTE|1||testing the system comments here|||
NTE|2||more comments here|||
OBR|1|ABC||||||||||||||1234^Acron^Smith~5678^Beta^Zeta|||||||||||||||||||||||||||||||||||||||||||
OBR|2|XYZ||||||||||||||903^Blacky|||||||||||||||||||||||||||||||||||||||||||`)

	d := NewDecoder(v25.Registry, &DecodeOption{
		HeaderOnly: true,
	})
	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 segment, got %d segments", len(v))
	}
	el := v[0]
	m, ok := el.(*v25.MSH)
	if !ok {
		t.Fatalf("expected MSG, got %T", el)
	}
	ck := func(name, g, w string) {
		if g != w {
			t.Errorf("for %s, got=<<%s>> want=<<%s>>", name, g, w)
		}
	}
	ck("security", m.Security, "XYZ")
	gr, err := d.DecodeGroup(v)
	mgr, ok := gr.(v25.ORL_O34)
	if !ok {
		t.Fatal("incorrect message type")
	}
	if err != nil {
		t.Error(err)
	}
	ck("group-security", mgr.MSH.Security, "XYZ")
	if mgr.Response != nil {
		t.Error("expected nil Response")
	}
}

func TestDecodeUnexpectedSegment(t *testing.T) {
	var raw = []byte(`MSH|^~\&|LAB|ORG|SYS||20250609071616||ORU^R01|1749478576661393532|P|2.5||||||UTF-8
MSA|AA|1749478576661393532|HL7
`)
	d := NewDecoder(v25.Registry, nil)
	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}

	_, err = group(v, v25.Registry)
	if err == nil {
		t.Fatal("expected err, got nil")
	}

	const want = `line 2 (*h250.MSA) not found in trigger "ORU_R01"`
	es := err.Error()
	if es != want {
		t.Fatalf("got: %q, want: %q", es, want)
	}
	var segmentError ErrUnexpectedSegment
	if !errors.As(err, &segmentError) {
		t.Fatalf("expected ErrUnexpectedSegment, got %T", err)
	}
	seg, ok := segmentError.Segment.(*v25.MSA)
	if !ok {
		t.Fatalf("wanted *h250.MSA got %T", segmentError)
	}
	if g, w := seg.AcknowledgmentCode, "AA"; g != w {
		t.Fatalf("got ack %q, wanted %q", g, w)
	}
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

func TestIgnoreRepetition(t *testing.T) {
	flag.Parse()
	var raw = []byte(`MSH|^~\&|MESA_OP|XYZ_HOSPITAL|iFW|ABC_HOSPITAL|20110613061611||SIU^S12|24916560|P|2.5||||||
SCH|10345^10345|2196178^2196178|||10345|OFFICE^Office visit|reason for the appointment|OFFICE|60|m|^^60^20110617084500^20110617093000|||||9^DENT^ARTHUR^||||9^DENT^COREY^|||||Scheduled
PID|1||42||SMITH^PAUL||19781012|M|||1 Broadway Ave^^Fort Wayne^IN^46804||(260)555-1234|||S||999999999|||||||||||||||||||||
PV1|1|O|||||1^Smith^Miranda^A^MD^^^^|2^Withers^Peter^D^MD^^^^||||||||||||||||||||||||||||||||||||||||||99158||
RGS|1|A
AIG|1|A|1^White, Charles~2^Black, Charles|D^^
AIL|1|A|OFFICE^^^OFFICE|^Main Office||20110614084500|||45|m^Minutes||Scheduled
AIP|1|A|1^White^Charles^A^MD^^^^|D^White, Douglas||20110614084500|||45|m^Minutes||Scheduled
`)

	d := NewDecoder(v25.Registry, &DecodeOption{
		IgnoreRepetition: true,
	})

	v, err := d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	g, err := d.DecodeGroup(v)
	if err != nil {
		t.Fatal(err)
	}
	if g == nil {
		t.Fatal("expected group, got nil")
	}

	d = NewDecoder(v25.Registry, &DecodeOption{})

	v, err = d.DecodeList(raw)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.DecodeGroup(v)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	const want = "line 6, AIG.ResourceID(*h250.CE)[3]: data repeats but element *h250.CE does not"
	errText := err.Error()
	if errText != want {
		t.Fatalf("got: %q, want: %q", errText, want)
	}

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

func TestMissingSep(t *testing.T) {
	var raw = `MSA|AA|undefined|HL7 ACK`
	d := NewDecoder(v25.Registry, nil)
	vv, err := d.DecodeList([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	var got string
	for _, v := range vv {
		switch v := v.(type) {
		case *v25.MSA:
			got = fmt.Sprintf("AC: %q-%q", v.AcknowledgmentCode, v.TextMessage)
		}
	}
	const want = `AC: "AA"-"HL7 ACK"`
	if got != want {
		t.Fatalf("got: %s\nwant: %s", got, want)
	}
}

func TestMissingSepDecode(t *testing.T) {
	var raw = `MSA|AA|undefined|HL7 ACK`
	d := NewDecoder(v25.Registry, nil)
	g, err := d.Decode([]byte(raw))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if g != nil {
		t.Fatalf("expected nil result, got: %v", g)
	}
	var seg ErrUnexpectedSegment
	isErr := errors.As(err, &seg)
	if !isErr {
		t.Fatalf("expected ErrUnexpectedSegment, got: %v", err)
	}
	v, ok := seg.Segment.(*v25.MSA)
	if !ok {
		t.Fatalf("expected *MSA, got: %T", seg.Segment)
	}
	got := fmt.Sprintf("AC: %q-%q", v.AcknowledgmentCode, v.TextMessage)

	const want = `AC: "AA"-"HL7 ACK"`
	if got != want {
		t.Fatalf("got: %s\nwant: %s", got, want)
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
				if err := os.WriteFile(fn, rt, 0600); err != nil {
					t.Fatal("overwrite", err)
				}
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
				if err := os.WriteFile(errorFn, gotError, 0600); err != nil {
					t.Fatal("overwrite", err)
				}
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
		// These are not properly formatted, but accept them anyway, the meaning is sufficiently clear.
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
		// The meaning is not sufficiently clear, reject.
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
