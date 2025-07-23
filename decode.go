package hl7

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type lineDecoder struct {
	sep       byte    // usually a |
	repeat    byte    // usually a ~
	dividers  [3]byte // usually |, ^, &
	chars     [4]byte // usually ^!\&
	escape    byte    // usually a \
	readSep   bool
	ignoreSep bool
	ignoreRep bool

	unescaper *strings.Replacer
}

// Decode bytes into HL7 structures.
type Decoder struct {
	registry Registry
	opt      DecodeOption
}

// Decode options for the HL7 decoder.
type DecodeOption struct {
	ErrorZSegment    bool // Error on an unknown Zxx segment when true.
	HeaderOnly       bool // Only decode first segment, usually the header.
	IgnoreFieldSep   bool // Ignore field separator values in text fields.
	IgnoreRepetition bool // Ignore repetitions in fields that are not repeatable.
}

// Create a new Decoder. A registry must be provided. Option is optional.
func NewDecoder(registry Registry, opt *DecodeOption) *Decoder {
	d := &Decoder{
		registry: registry,
	}
	if opt != nil {
		d.opt = *opt
	}

	return d
}

// Decode takes an hl7 message and returns a final trigger with all segments grouped.
func (d *Decoder) Decode(data []byte) (any, error) {
	list, err := d.DecodeList(data)
	if err != nil {
		return nil, fmt.Errorf("segment list: %w", err)
	}
	g, err := d.DecodeGroup(list)
	if err != nil {
		return nil, fmt.Errorf("trigger group: %w", err)
	}
	return g, nil
}

// Group a list of elements into trigger groupings.
// A value and error may be present at the same time.
func (d *Decoder) DecodeGroup(list []any) (any, error) {
	return group(list, d.registry)
}

// Varies should be implemented on a segment that knows how to
// decode a child VARIES data type.
type Varies interface {
	ChildVaries(reg func(string) (any, bool)) (reflect.Value, error)
}

type variesFunc func() (reflect.Value, error)

var variesType = reflect.TypeFor[Varies]()

// SegmentError may be returned as part of the DecodeList result.
// This allows a single segment to be decoded poorly with error, while
// still decoding the rest of the message.
// This will not be returned as an error.
type SegmentError struct {
	ErrorList []error
	Segment   any
}

func (e SegmentError) Error() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "errors in segment %[1]T=%[1]v:", e.Segment)
	for _, err := range e.ErrorList {
		sb.WriteRune('\n')
		sb.WriteRune('\t')
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func (e SegmentError) Unwrap() []error {
	return e.ErrorList
}

// Decode returns a list of segments without any grouping applied.
func (d *Decoder) DecodeList(data []byte) ([]any, error) {
	// Explicitly accept both CR and LF as new lines. Some systems do use \n, despite the spec.
	lines := bytes.FieldsFunc(data, func(r rune) bool {
		switch r {
		default:
			return false
		case '\r', '\n':
			return true
		}
	})

	type field struct {
		name  string
		index int
		tag   tag
		field reflect.Value
	}

	ret := []any{}

	ld := &lineDecoder{
		ignoreSep: d.opt.IgnoreFieldSep,
		ignoreRep: d.opt.IgnoreRepetition,
	}
	for index, line := range lines {
		lineNumber := index + 1
		if len(line) == 0 {
			continue
		}

		segTypeName, n := ld.getID(line)
		remain := line[n:]
		if len(segTypeName) == 0 {
			return nil, fmt.Errorf("line %d: missing segment type", lineNumber)
		}

		seg, ok := d.registry.Segment(segTypeName)
		if !ok {
			isZ := len(segTypeName) > 0 && segTypeName[0] == 'Z'
			if isZ && !d.opt.ErrorZSegment {
				continue
			}
			return nil, fmt.Errorf("line %d: unknown segment type %q", lineNumber, segTypeName)
		}

		rt := reflect.TypeOf(seg)
		ct := rt.NumField()

		fieldList := make([]field, 0, ct)

		hasInit := false

		rv := reflect.New(rt)
		rvv := rv.Elem()

		var SegmentName string
		var SegmentSize int32
		var maxOrd int32

		for i := 0; i < ct; i++ {
			ft := rt.Field(i)
			tagText := ft.Tag.Get(tagName)
			tag, err := parseTag(ft.Name, tagText)
			if err != nil {
				return nil, err
			}
			if !tag.Present {
				continue
			}
			if tag.Meta {
				SegmentName = tag.Name
				SegmentSize = tag.Order
				if ft.Type.Kind() == reflect.String {
					rvv.Field(i).SetString(tag.Name)
				}
				continue
			}
			if tag.Order > maxOrd {
				maxOrd = tag.Order
			}
			if tag.FieldSep || tag.FieldChars {
				hasInit = true
			}
			f := field{
				name:  ft.Name,
				index: i,
				tag:   tag,
			}
			f.field = rvv.Field(i)

			if !f.field.IsValid() {
				return nil, fmt.Errorf("%s.%s invalid reflect value", SegmentName, f.name)
			}

			fieldList = append(fieldList, f)
		}
		if SegmentSize == 0 {
			SegmentSize = maxOrd
		}
		SegmentFieldLength := int(SegmentSize + 1)

		offset := 0
		if hasInit {
			if len(remain) < 5 {
				return nil, fmt.Errorf("missing format delims")
			}
			ld.sep = remain[0]
			copy(ld.chars[:], remain[1:5])

			ld.dividers = [3]byte{ld.sep, ld.chars[0], ld.chars[3]}
			ld.repeat = ld.chars[1]
			ld.escape = ld.chars[2]
			ld.setupUnescaper()
			ld.readSep = true

			remain = remain[5:]
			offset = 2
		}

		if ld.sep == 0 {
			return nil, fmt.Errorf("missing sep prior to field")
		}

		parts := bytes.Split(remain, []byte{ld.sep})

		ff := make([]field, SegmentFieldLength)
		for _, f := range fieldList {
			if f.tag.FieldSep {
				f.field.SetString(string(ld.sep))
				continue
			}
			if f.tag.FieldChars {
				f.field.SetString(string(ld.chars[:]))
				continue
			}
			index := int(f.tag.Order) - offset
			if index < 0 || index >= SegmentFieldLength {
				continue
			}
			ff[index] = f
		}

		var vfc variesFunc
		if rvv.Type().Implements(variesType) {
			vfc = func() (reflect.Value, error) {
				return rvv.Interface().(Varies).ChildVaries(d.registry.DataType)
			}
		}

		var segmentErrorList []error
		for i, f := range ff {
			if i >= len(parts) {
				break
			}
			p := parts[i]
			if !f.tag.Present {
				continue
			}
			if f.tag.Omit {
				continue
			}
			err := ld.decodeSegmentList(p, f.tag, f.field, vfc)
			if err != nil {
				if v, ok := err.(*DecodeSegmentError); ok && v.Line == 0 && len(v.SegmentName) == 0 && len(v.FieldName) == 0 {
					v.Line = lineNumber
					v.SegmentName = SegmentName
					v.FieldName = f.name
				} else {
					err = &DecodeSegmentError{
						Line:        lineNumber,
						SegmentName: SegmentName,
						FieldName:   f.name,
						Inner:       err,
					}
				}
				segmentErrorList = append(segmentErrorList, err)
			}
		}
		if segmentErrorList != nil {
			ret = append(ret, SegmentError{
				ErrorList: segmentErrorList,
				Segment:   rv.Interface(),
			})
		} else {
			ret = append(ret, rv.Interface())
		}
		if d.opt.HeaderOnly {
			return ret, nil
		}
	}
	return ret, nil
}

func (d *lineDecoder) setupUnescaper() {
	d.unescaper = strings.NewReplacer(
		string([]byte{d.escape, 'F', d.escape}), string(d.sep),
		string([]byte{d.escape, 'S', d.escape}), string(d.chars[0]),
		string([]byte{d.escape, 'R', d.escape}), string(d.chars[1]),
		string([]byte{d.escape, 'E', d.escape}), string(d.chars[2]),
		string([]byte{d.escape, 'T', d.escape}), string(d.chars[3]),
	)
}

var timeType reflect.Type = reflect.TypeOf(time.Time{})

func (d *lineDecoder) decodeSegmentList(data []byte, t tag, rv reflect.Value, vfc variesFunc) error {
	if len(data) == 0 {
		return nil
	}
	parts := bytes.Split(data, []byte{d.repeat})
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}
		var err error
		if d.ignoreRep && len(parts) > 1 && rv.Kind() != reflect.Slice {
			// Decode only the first repetition and ignore the rest.
			err = d.decodeSegment(p, t, rv, 1, false, vfc)
			if err == nil {
				// If we successfully decoded the first repetition, we can break out.
				break
			}
		} else {
			err = d.decodeSegment(p, t, rv, 1, len(parts) > 1, vfc)
		}
		if err != nil {
			return &DecodeSegmentError{
				FieldType: rv.Type().String(),
				Ordinal:   t.Order,
				Inner:     err,
			}
		}
	}
	return nil
}

type DecodeSegmentError struct {
	Line        int
	SegmentName string
	FieldName   string
	FieldType   string
	Ordinal     int32
	Inner       error
}

func (e *DecodeSegmentError) Error() string {
	sb := &strings.Builder{}
	if e.Line > 0 {
		sb.WriteString("line ")
		sb.WriteString(strconv.FormatInt(int64(e.Line), 10))
		sb.WriteString(", ")
	}
	if len(e.SegmentName) > 0 {
		sb.WriteString(e.SegmentName)
		sb.WriteString(".")
	}
	if len(e.FieldName) > 0 {
		sb.WriteString(e.FieldName)
	}
	if len(e.FieldType) > 0 {
		sb.WriteRune('(')
		sb.WriteString(e.FieldType)
		sb.WriteRune(')')
	}
	if e.Ordinal > 0 {
		sb.WriteRune('[')
		sb.WriteString(strconv.FormatInt(int64(e.Ordinal), 10))
		sb.WriteRune(']')
	}
	if e.Inner != nil {
		sb.WriteString(": ")
		sb.WriteString(e.Inner.Error())
	}
	return sb.String()
}
func (e *DecodeSegmentError) Unwrap() error {
	return e.Inner
}

type DecodeDataError struct {
	Tag   string
	Value string
}

func (e *DecodeDataError) Error() string {
	return fmt.Sprintf("%s contains an escape character %s; data may be malformed, invalid type, or contain a bug", e.Tag, e.Value)
}

func (d *lineDecoder) decodeSegment(data []byte, t tag, rv reflect.Value, level int, mustBeSlice bool, vfc variesFunc) error {
	type field struct {
		tag   tag
		field reflect.Value
	}

	isSlice := rv.Kind() == reflect.Slice
	if mustBeSlice && !isSlice {
		return fmt.Errorf("data repeats but element %v does not", rv.Type())
	}

	switch rv.Kind() {
	default:
		return fmt.Errorf("unknown field kind %v value=%v(%v) tag=%v data=%q", rv.Kind(), rv, rv.Type(), t, data)
	case reflect.Interface:
		if vfc == nil {
			return fmt.Errorf("unsupported interface field kind %#v data=%q", t, data)
		}
		nextRV, err := vfc()
		if err != nil {
			return err
		}
		err = d.decodeSegment(data, t, nextRV, level, mustBeSlice, vfc)
		rv.Set(nextRV)
		return err
	case reflect.Pointer:
		next := reflect.New(rv.Type().Elem())
		rv.Set(next)
		return d.decodeSegment(data, t, next.Elem(), level, false, vfc)
	case reflect.Slice:
		if len(data) == 0 {
			return nil
		}
		itemType := rv.Type().Elem()
		if itemType.Kind() == reflect.Uint8 {
			rv.SetBytes(data)
			return nil
		}
		itemValue := reflect.New(itemType)
		ivv := itemValue.Elem()
		err := d.decodeSegment(data, t, ivv, level, false, vfc)
		if err != nil {
			return fmt.Errorf("slice: %w", err)
		}

		rv.Set(reflect.Append(rv, ivv))
		return nil
	case reflect.Struct:
		switch rv.Type() {
		default:
			sep := d.dividers[level]

			rt := rv.Type()
			ct := rv.NumField()

			fieldList := []field{}

			var SegmentName string
			var SegmentSize int32
			var maxOrd int32

			for i := 0; i < ct; i++ {
				ft := rt.Field(i)
				fTag, err := parseTag(ft.Name, ft.Tag.Get(tagName))
				if err != nil {
					return err
				}

				if fTag.Meta {
					SegmentName = fTag.Name
					SegmentSize = fTag.Order
					if ft.Type.Kind() == reflect.String {
						rv.Field(i).SetString(SegmentName)
					}
					continue
				}
				if !fTag.Present {
					continue
				}
				if fTag.Omit {
					continue
				}
				if fTag.Order > maxOrd {
					maxOrd = fTag.Order
				}

				fv := rv.Field(i)

				f := field{
					tag:   fTag,
					field: fv,
				}
				fieldList = append(fieldList, f)
			}
			if SegmentSize == 0 {
				SegmentSize = maxOrd
			}
			ff := make([]field, int(SegmentSize))

			for _, f := range fieldList {
				index := int(f.tag.Order - 1)
				if index < 0 || index >= len(ff) {
					continue
				}

				ff[index] = f
			}

			// TODO: Make more robust. Watch for repeats, etc, other stuff.
			parts := bytes.Split(data, []byte{sep})
			for i, p := range parts {
				if i >= len(ff) {
					continue
				}
				f := ff[i]
				err := d.decodeSegment(p, f.tag, f.field, level+1, false, vfc)
				if err != nil {
					return &DecodeSegmentError{
						SegmentName: SegmentName,
						FieldType:   f.field.Type().String(),
						FieldName:   f.tag.Name,
						Ordinal:     f.tag.Order,
						Inner:       err,
					}
				}
			}
			return nil
		case timeType:
			v := d.decodeByte(data, t)
			t, err := d.parseDateTime(v)
			if err != nil {
				return err
			}
			rv.Set(reflect.ValueOf(t))
			return nil
		}
	case reflect.String:
		c1, c2, c3 := d.dividers[0], d.dividers[1], d.dividers[2]
		if !d.ignoreSep {
			for _, b := range data {
				switch b {
				case c1, c2, c3:
					return &DecodeDataError{
						Tag:   t.Name,
						Value: string(data),
					}
				}
			}
		}
		rv.SetString(d.decodeByte(data, t))
		return nil
	}
}
func (d *lineDecoder) decodeByte(v []byte, t tag) string {
	if t.NoEscape {
		return string(v)
	}
	return d.unescaper.Replace(string(v))
}

func (d *lineDecoder) getID(data []byte) (string, int) {
	if d.readSep {
		v, _, _ := bytes.Cut(data, []byte{d.sep})
		return string(v), len(v)
	}
	for i, r := range data {
		if unicode.IsLetter(rune(r)) || unicode.IsNumber(rune(r)) {
			continue
		}
		return string(data[:i]), i
	}
	return string(data), len(data)
}

func dtClean(s string) string {
	var i int
	return strings.Map(func(r rune) rune {
		i++
		switch r {
		default:
			return r
		case ' ', ':':
			return -1
		case '-':
			if i <= 8 {
				return -1
			}
			return r
		}
	}, s)
}

func (d *lineDecoder) parseDateTime(dt string) (time.Time, error) {
	var zoneIndex int
	dt = dtClean(dt)
	dtLen := len(dt)
loop:
	for i, r := range dt {
		switch {
		default:
			return time.Time{}, fmt.Errorf("invalid characters in date: %q", dt)
		case unicode.IsNumber(r):
		case r == '.':
		case r == '-':
			zoneIndex = i
		case r == '+':
			zoneIndex = i
		case r == '^':
			dtLen = i
			break loop
		}
	}
	dt = dt[:dtLen]

	// Format: YYYY[MM[DD[HH[MM[SS[.S[S[S[S]]]]]]]]][+/-ZZZZ]^<degree of precision>
	// 20200522143859198-0700
	// 20060102150405
	in := dt
	var t time.Time
	var err error
	if zoneIndex > 0 {
		tp := dt[:zoneIndex]
		zp := dt[zoneIndex:]

		switch len(dt) {
		default:
			if len(tp) < 12 {
				return t, fmt.Errorf("unknown date time string size %q", tp)
			}
			in = tp[:12] + zp
			t, err = time.Parse("200601021504-0700", in)
		case 4 + 5: // Year.
			t, err = time.Parse("2006-0700", in)
		case 6 + 5: // Month.
			t, err = time.Parse("2006-0700", in)
		case 8 + 5: // To the day.
			t, err = time.Parse("20060102-0700", in)
		case 12 + 5: // To the minute.
			t, err = time.Parse("200601021504-0700", in)
		case 14 + 5, 16 + 5: // To the second.
			in = tp[:14] + zp
			t, err = time.Parse("20060102150405-0700", in)
		}
		if err != nil {
			err = fmt.Errorf("field %q: %w", dt, err)
		}
		return t, err
	}
	switch len(dt) {
	default:
		if len(dt) < 12 {
			return t, fmt.Errorf("unknown date time string size %q", dt)
		}
		in = dt[:12]
		t, err = time.Parse("200601021504", in)
	case 0:
		t, err = time.Time{}, nil // No date supplied, use zero value
	case 4: // Year.
		t, err = time.Parse("2006", in)
	case 6: // Month
		t, err = time.Parse("200601", in)
	case 8: // To the day.
		t, err = time.Parse("20060102", in)
	case 12: // To the minute.
		t, err = time.Parse("200601021504", in)
	case 14, 16: // To the second.
		in = dt[:14]
		t, err = time.Parse("20060102150405", in)
	}
	return t, err
}
