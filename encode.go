package hl7

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

const nextLine = '\r'
const defaultSep = "|"
const defaultChars = `^~\&`

// Create a new Encoder. Options may be nil.
func NewEncoder(opt *EncodeOption) *Encoder {
	e := &Encoder{}
	if opt != nil {
		e.opt = *opt
	}
	return e
}

// Encoding options.
type EncodeOption struct {
	TrimTrailingSeparator bool
}

type Encoder struct {
	// Set only from init.
	initSep   string
	initChars string

	sep      byte // usually a |
	repeat   byte // usually a ~
	dividers []byte
	esc      map[byte][]byte

	deferred [3]*bytes.Buffer
	buf      *bytes.Buffer

	opt EncodeOption
}

func (e *Encoder) Encode(message any) ([]byte, error) {
	e.init("", "")

	err := e.walk(1, reflect.ValueOf(message))
	if err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}

// Init separators and reset buffers.
// If sep or chars are empty, then the previous value or the default will be used.
func (e *Encoder) init(sep, chars string) {
	if len(sep) > 0 {
		sep = e.initSep
	}
	if len(sep) == 0 {
		sep = defaultSep
	}

	if len(chars) == 0 {
		chars = e.initSep
	}
	if len(chars) == 0 {
		chars = defaultChars
	}

	e.sep = byte(sep[0])
	e.repeat = chars[1]
	e.dividers = []byte{sep[0], chars[0], chars[3]}
	if e.deferred[0] == nil {
		e.deferred[0] = &bytes.Buffer{}
	}
	if e.deferred[1] == nil {
		e.deferred[1] = &bytes.Buffer{}
	}
	if e.deferred[2] == nil {
		e.deferred[2] = &bytes.Buffer{}
	}
	if e.buf == nil {
		e.buf = &bytes.Buffer{}
	}
	e.buf.Reset()
	for _, d := range e.deferred {
		d.Reset()
	}

	esc := chars[2]
	e.esc = map[byte][]byte{
		e.sep:    {esc, 'F', esc},
		chars[0]: {esc, 'S', esc},
		chars[1]: {esc, 'R', esc},
		chars[2]: {esc, 'E', esc},
		chars[3]: {esc, 'T', esc},
	}
}

func (e *Encoder) walk(seq int, wv reflect.Value) error {
	if !wv.IsValid() {
		return nil
	}
	metaTag, err := e.meta(wv.Type())
	if err != nil {
		return err
	}
	if !metaTag.Present {
		return nil
	}
	switch metaTag.Type {
	default:
		return nil
	case structTrigger, structTriggerGroup:
		switch wv.Kind() {
		default:
			return nil
		case reflect.Pointer:
			return e.walk(seq, wv.Elem())
		case reflect.Slice:
			l := wv.Len()
			for i := 0; i < l; i++ {
				v := wv.Index(i)
				err := e.walk(i+1, v)
				if err != nil {
					return err
				}
			}
			return nil
		case reflect.Struct:
			ct := wv.NumField()
			for i := 0; i < ct; i++ {
				v := wv.Field(i)
				err := e.walk(seq, v)
				if err != nil {
					return err
				}
			}
			return nil
		}
	case structSegment:
		switch wv.Kind() {
		default:
			return nil
		case reflect.Pointer:
			return e.walk(seq, wv.Elem())
		case reflect.Slice:
			l := wv.Len()
			for i := 0; i < l; i++ {
				v := wv.Index(i)
				err := e.encodeSegment(i+1, v)
				if err != nil {
					return err
				}
			}
			return nil
		case reflect.Struct:
			return e.encodeSegment(seq, wv)
		}
	}
}
func (e *Encoder) meta(wt reflect.Type) (tag, error) {
	switch wt.Kind() {
	default:
		return tag{}, nil
	case reflect.Pointer:
		return e.meta(wt.Elem())
	case reflect.Slice:
		return e.meta(wt.Elem())
	case reflect.Struct:
		sf, ok := wt.FieldByName(hl7MetaName)
		if !ok {
			return tag{}, nil
		}
		return parseTag(sf.Name, sf.Tag.Get(tagName))
	}
}

// encodeSegment the given message into buffer.
func (e *Encoder) encodeSegment(seq int, st reflect.Value) error {
	stt := st.Type()

	var SegmentName string
	var SegmentSize int32
	var maxOrd int32

	type field struct {
		name    string
		present bool
		tag     tag
		value   any
	}
	var fieldList []field

	var msgSep string
	for i := 0; i < st.NumField(); i++ {
		fld := stt.Field(i)
		f := st.Field(i)
		tag, err := parseTag(fld.Name, fld.Tag.Get(tagName))
		if err != nil {
			return err
		}
		switch {
		case tag.FieldSep:
			msgSep = f.String()
		case tag.FieldChars:
			chars := f.String()
			e.init(msgSep, chars)
		}

		if tag.Meta {
			switch tag.Type {
			case structTrigger, structTriggerGroup:
				return fmt.Errorf("trigger and trigger group structures should not be passed in to encode, package error")
			}
		}
		if !tag.Present {
			continue
		}
		if tag.Meta {
			SegmentName = tag.Name
			SegmentSize = tag.Order
		} else {
			if tag.Order > maxOrd {
				maxOrd = tag.Order
			}
		}

		fieldList = append(fieldList, field{
			name:    fld.Name,
			present: !f.IsZero(),
			tag:     tag,
			value:   f.Interface(),
		})
	}

	if SegmentSize == 0 {
		SegmentSize = maxOrd
	}
	ff := make([]field, SegmentSize)
	for _, f := range fieldList {
		index := f.tag.Order - 1
		if index < 0 || index >= SegmentSize {
			continue
		}
		ff[index] = f
	}

	e.write(SegmentName, 0, true)
	for _, f := range ff {
		if f.tag.Omit {
			continue
		}
		v := f.value
		switch {
		// Auto populate SetID sequences.
		case f.tag.Sequence:
			if seq == 0 {
				return fmt.Errorf("incorrect zero sequence number")
			}
			if s, ok := v.(string); ok && len(s) == 0 {
				v = strconv.FormatInt(int64(seq), 10)
			}
		}
		direct := !e.opt.TrimTrailingSeparator
		e.writeSep(0, 0, direct)
		err := e.encodeDataType(f.tag, v, 0)
		if err != nil {
			return err
		}
	}
	e.resetAllDeferred()
	e.writeSep(0, nextLine, false)
	return nil
}

func (e *Encoder) flushDeferred(level int) {
	// If level 0"|", then write level 0, remove 1, 2.
	// If level 1"^", then write level 0 and 1, remove 2.
	// If level 2"&", then write level 0, 1, and 2.
	for index, d := range e.deferred {
		if index <= level {
			db := d.Bytes()
			e.buf.Write(db)
		}
		d.Reset()
	}
}

// write escapes individual values.
func (e *Encoder) write(val string, level int, noEscape bool) {
	if len(val) > 0 {
		e.flushDeferred(level)
	}
	buf := e.buf
	if noEscape {
		buf.WriteString(val)
		return
	}
	for i := 0; i < len(val); i++ {
		c := val[i]
		if esc, is := e.esc[c]; is {
			buf.Write(esc)
			continue
		}
		buf.WriteByte(c)
	}
}
func (e *Encoder) resetAllDeferred() {
	for _, d := range e.deferred {
		d.Reset()
	}
}
func (e *Encoder) writeSep(level int, sep byte, direct bool) {
	if sep == 0 {
		sep = e.dividers[level]
	}
	// If level 0"|", then write it directly and reset the other layers.
	// If level 1"&", then reset 2 and write to defered 1.
	for index, d := range e.deferred {
		if index > level {
			d.Reset()
		}
	}
	if direct {
		e.deferred[level].WriteByte(sep)
		e.flushDeferred(level)
		return
	}
	e.deferred[level].WriteByte(sep)
}

func (e *Encoder) encodeDataType(t tag, o interface{}, level int) error {
	if o == nil || !t.Present {
		return nil
	}

	switch v := o.(type) {
	default:
		rv := reflect.ValueOf(o)
		if rv.IsZero() {
			return nil
		}
		switch rv.Kind() {
		default:
			return fmt.Errorf("unknown value kind: %v", rv.Kind())
		case reflect.Pointer:
			rv = rv.Elem()
			fallthrough
		case reflect.Struct:
			var SegmentName string
			var SegmentSize int32
			var maxOrd int32

			type field struct {
				name    string
				present bool
				tag     tag
				value   any
			}
			var fieldList []field

			rt := rv.Type()
			ct := rt.NumField()

			for i := 0; i < ct; i++ {
				ft := rt.Field(i)
				tag, err := parseTag(ft.Name, ft.Tag.Get(tagName))
				if err != nil {
					return err
				}

				if !tag.Present {
					continue
				}
				if tag.Meta {
					SegmentName = tag.Name
					SegmentSize = tag.Order
				} else {
					if tag.Order > maxOrd {
						maxOrd = tag.Order
					}
				}
				f := rv.Field(i)

				fieldList = append(fieldList, field{
					name:    ft.Name,
					present: !f.IsZero(),
					tag:     tag,
					value:   f.Interface(),
				})
			}

			if SegmentSize == 0 {
				SegmentSize = maxOrd
			}
			ff := make([]field, SegmentSize)
			for _, f := range fieldList {
				index := f.tag.Order - 1
				if index < 0 || index >= SegmentSize {
					continue
				}
				ff[index] = f
			}

			for i, f := range ff {
				if i != 0 {
					e.writeSep(level+1, 0, false)
				}
				err := e.encodeDataType(f.tag, f.value, level+1)
				if err != nil {
					return fmt.Errorf("%s (%+v): %w", SegmentName, f.value, err)
				}
			}
		case reflect.Slice:
			ct := rv.Len()
			for i := 0; i < ct; i++ {
				if i != 0 {
					e.writeSep(level, e.repeat, true)
				}
				x := rv.Index(i)
				value := x.Interface()

				err := e.encodeDataType(t, value, level)
				if err != nil {
					return fmt.Errorf("repeat (%+v): %w", value, err)
				}
			}
			return nil
		}
	case string:
		e.write(v, level, t.NoEscape)
	case time.Time:
		if v.IsZero() {
			return nil
		}
		var sv string
		switch t.Format {
		default:
			sv = v.Format("20060102150405")
		case "YMDHMS":
			sv = v.Format("20060102150405")
		case "YMDHM":
			sv = v.Format("200601021504")
		case "YMD":
			sv = v.Format("20060102")
		case "HM":
			sv = v.Format("1504")
		}
		e.write(sv, level, t.NoEscape)
	}
	return nil
}
