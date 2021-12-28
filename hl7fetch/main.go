package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/kardianos/task"
)

//go:embed template/*.go.template
var pkgTemplate embed.FS

/*
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Chapters
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.8/Chapters/CH_02
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/TriggerEvents
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/TriggerEvents/ORU_R30
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Segments/NK1
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/DataTypes/CE
	https://hl7-definition.caristix.com/v2-api/1/HL7v2.5/Tables/0052
*/
const rootURL = `https://hl7-definition.caristix.com/v2-api/1/`

func main() {
	err := task.Start(context.Background(), time.Second*2, run)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	versionList := []struct {
		text  string
		value string
	}{
		{
			text:  "HL7 v2.1",
			value: "HL7v2.1",
		}, {
			text:  "HL7 v2.2",
			value: "HL7v2.2",
		}, {
			text:  "HL7 v2.3",
			value: "HL7v2.3",
		}, {
			text:  "HL7 v2.3.1",
			value: "HL7v2.3.1",
		}, {
			text:  "HL7 v2.4",
			value: "HL7v2.4",
		}, {
			text:  "HL7 v2.5",
			value: "HL7v2.5",
		}, {
			text:  "HL7 v2.5.1",
			value: "HL7v2.5.1",
		}, {
			text:  "HL7 v2.6",
			value: "HL7v2.6",
		}, {
			text:  "HL7 v2.7",
			value: "HL7v2.7",
		}, {
			text:  "HL7 v2.7.1",
			value: "HL7v2.7.1",
		}, {
			text:  "HL7 v2.8",
			value: "HL7v2.8",
		},
	}

	versionPrefix := "HL7v"

	version := flag.String("version", "2.5", "Version of Data")
	rootDir := flag.String("root", "", "root of data")
	pkgDir := flag.String("pkgdir", "", "package directory for generated files")
	allowNetwork := flag.Bool("network", false, "allow making network calls")
	flag.Parse()

	if len(*rootDir) == 0 {
		return fmt.Errorf("missing root flag")
	}
	knownVersion := false
	apiVersion := versionPrefix + *version
	for _, v := range versionList {
		if apiVersion == v.value {
			knownVersion = true
			break
		}
	}

	if !knownVersion {
		list := &strings.Builder{}
		for _, v := range versionList {
			list.WriteString(v.value)
			list.WriteString(" => ")
			list.WriteString(v.text)
			list.WriteString("\n")
		}
		return fmt.Errorf("unknown version %q, all known versions are:\n%s", *version, list.String())
	}

	r := &runner{
		version:      apiVersion,
		rootDir:      *rootDir,
		allowNetwork: *allowNetwork,
		pkgDir:       *pkgDir,

		cache: make(map[cacheKey]any, 100),

		client: &http.Client{
			Transport: &roundTripper{
				UA:        `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36`,
				Transport: http.DefaultTransport.(*http.Transport),
				MinDelay:  time.Millisecond * 400,
				Header: map[string]string{
					"Connection":                "keep-alive",
					"Cache-Control":             "max-age=0",
					"Upgrade-Insecure-Requests": "1",
					"Accept":                    "application/json, text/plain, */*",
				},
			},
		},
	}

	// HL7 2-8.
	controlSegments := []string{
		"BHS",
		"BTS",
		"FHS",
		"FTS",
		"DSC",
		"OVR",
		"ADD",
		"SFT",
		"SGT",
		"ARV",
		"UAC",
	}

	// Go through all items to pull them into the cache.
	tlist, err := r.getTriggerList()
	if err != nil {
		return err
	}
	visitSement := func(name string) error {
		if len(name) == 0 {
			return nil
		}
		seg, err := r.getSegment(name)
		if err != nil {
			return err
		}

		for _, f := range seg.Fields {
			seen := make(map[string]bool, 10)
			err = visitAllDataType(r, seen, f.DataType, func(dt DataType) error {
				for _, f := range dt.Fields {
					if len(f.TableID) > 0 {
						_, err := r.getTable(f.TableID)
						if err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			if len(f.TableID) > 0 {
				_, err := r.getTable(f.TableID)
				if err != nil {
					return err
				}
			}
		}
		return err
	}
	for _, item := range tlist {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		trigger, err := r.getTrigger(item.ID)
		if err != nil {
			return err
		}

		for _, ch := range trigger.Chapters {
			_, err = r.getChapter(ch)
			if err != nil {
				return err
			}
		}

		err = visitAllSegment("", nil, trigger.Segments, func(_ string, _, s *TriggerSegment) error {
			err := visitSement(s.ID)
			if err != nil && errors.Is(err, errIgnore) {
				return nil
			}
			return err
		})
		if err != nil {
			return err
		}
	}
	printCtl := []string{}
	for _, name := range controlSegments {
		err = visitSement(name)
		if errors.Is(err, errIgnore) {
			continue
		}
		if err != nil {
			return err
		}
		printCtl = append(printCtl, name)
	}

	if len(r.pkgDir) == 0 {
		return nil
	}

	pkgName := func(fn string) string {
		parts := strings.Split(fn, string(filepath.Separator))
		return strings.ToLower(parts[len(parts)-1])
	}(r.pkgDir)

	type To struct {
		Command     string
		PackageName string
		HL7Version  string
		Chapter     []Chapter
		Trigger     []TriggerType
		Segment     []SegmentType
		DataType    []DataType
		Table       []TableType

		ControlSegment []string
	}

	to := To{
		Command:     "hl7fetch " + strings.Join(os.Args[1:], " "),
		PackageName: pkgName,
		HL7Version:  *version,

		ControlSegment: printCtl,
	}
	for _, v := range r.cache {
		switch v := v.(type) {
		case Chapter:
			to.Chapter = append(to.Chapter, v)
		case TriggerType:
			to.Trigger = append(to.Trigger, v)
		case SegmentType:
			to.Segment = append(to.Segment, v)
		case DataType:
			to.DataType = append(to.DataType, v)
		case TableType:
			to.Table = append(to.Table, v)
		}
	}
	sort.Slice(to.Chapter, func(i, j int) bool {
		a, b := to.Chapter[i], to.Chapter[j]
		return a.ID < b.ID
	})
	sort.Slice(to.Trigger, func(i, j int) bool {
		a, b := to.Trigger[i], to.Trigger[j]
		return a.ID < b.ID
	})
	sort.Slice(to.Segment, func(i, j int) bool {
		a, b := to.Segment[i], to.Segment[j]
		return a.ID < b.ID
	})
	sort.Slice(to.DataType, func(i, j int) bool {
		a, b := to.DataType[i], to.DataType[j]
		return a.ID < b.ID
	})
	sort.Slice(to.Table, func(i, j int) bool {
		a, b := to.Table[i], to.Table[j]
		return a.ID < b.ID
	})

	displaySan := strings.NewReplacer(
		",", "-",
		`"`, `'`,
		"`", `'`,
		`=`, `:`,
		"\r\n", ` `,
		"\n", ` `,
		"\r", ` `,
	)

	rawMap := func(r rune) rune {
		if r == '`' {
			return '\''
		}
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}
	idMap := func(r rune) rune {
		switch {
		default:
			return -1
		case unicode.IsLetter(r):
			return r
		case unicode.IsDigit(r):
			return r
		}
	}

	processField := func(containerID string, f *Field, ord int, unique map[string]int) {
		name := f.Name

		name, _, _ = strings.Cut(name, "(e.g.")
		name = strings.Map(func(r rune) rune {
			if r == '\'' {
				return -1
			}
			if unicode.IsSymbol(r) || unicode.IsPunct(r) {
				return ' '
			}
			if !unicode.IsLetter(r) && !unicode.IsNumber(r) && !unicode.IsSpace(r) {
				return -1
			}
			return r
		}, name)
		ordS := strconv.FormatInt(int64(ord), 10)
		parts := strings.Fields(name)
		ret := make([]string, 0, len(parts))
		for _, v := range parts {
			omit := false
			switch {
			default:
				v = strings.Title(v)
			case strings.EqualFold(v, "id"):
				v = strings.ToUpper(v)
			case strings.EqualFold(v, "cpu"):
				v = strings.ToUpper(v)
			case strings.EqualFold(v, containerID):
				omit = true
			case ordS == v:
				omit = true
			}
			if !omit && len(v) > 0 {
				ret = append(ret, v)
			}
		}
		var id string
		if len(ret) == 0 {
			id = "Value"
		} else {
			id = strings.Join(ret, "")
		}
		ct := unique[id]
		ct++
		unique[id] = ct
		if ct > 1 {
			id = id + strconv.FormatInt(int64(ct), 10)
		}
		f.ID = id
	}

	emptyDataTypes := map[string]bool{}
	for i := range to.DataType {
		unique := map[string]int{}
		dt := &to.DataType[i]
		if len(dt.Fields) == 0 {
			emptyDataTypes[dt.ID] = true
		}
		for j := range dt.Fields {
			f := &dt.Fields[j]
			processField(dt.ID, f, j+1, unique)
		}
	}
	for i := range to.Segment {
		unique := map[string]int{}
		s := &to.Segment[i]
		for j := range s.Fields {
			f := &s.Fields[j]
			processField(s.ID, f, j+1, unique)
		}
	}

	dataTypeGlobalUnique := map[string]int{}
	for _, x := range to.Trigger {
		uniquePerStruct := map[*TriggerSegment]map[string]int{}
		// 1. Normalize group name.
		// 2. Give each group an ID top + ID (data type.)
		visitAllSegment(x.ID, nil, x.Segments, func(top string, parent, s *TriggerSegment) error {
			unique, ok := uniquePerStruct[parent]
			if !ok {
				unique = make(map[string]int, 10)
				uniquePerStruct[parent] = unique
			}
			if s.IsGroup {
				id := strings.ToLower(s.Name)
				id = strings.Title(id)
				if strings.HasSuffix(id, "Id") {
					id = id[:len(id)-2] + "ID"
				}
				s.LongName = id
				id = strings.Map(idMap, id)
				s.ID = id               // Field Name
				s.Name = top + "_" + id // Data type

				// Ensure new data type is unique.
				key := s.Name
				if len(key) == 0 {
					panic("empty key on visit: " + top)
				}
				ct := dataTypeGlobalUnique[key]
				dataTypeGlobalUnique[key] = ct + 1
				if ct == 0 {
					return nil
				}
				s.Name = s.Name + strconv.FormatInt(int64(ct+1), 10)
			}
			key := s.ID
			if len(key) == 0 {
				panic("empty key on visit: " + top)
			}
			ct := unique[key]
			unique[key] = ct + 1
			if ct == 0 {
				return nil
			}
			s.ID = s.ID + strconv.FormatInt(int64(ct+1), 10)
			return nil
		})
	}

	type memorizeKey struct {
		id   string
		name string
	}
	type memorizeValue struct {
		ID       string
		Name     string
		Segments []TriggerSegment
	}
	memorizeUnique := map[memorizeKey]bool{}
	var memorizeList []memorizeValue

	funcs := map[string]any{
		"raw": func(c string) string {
			return strings.Map(rawMap, c)
		},
		"comment": func(c string) string {
			if len(c) == 0 {
				return ""
			}
			buf := &strings.Builder{}
			ct := 0
			for _, r := range c {
				switch r {
				default:
					ct++
					buf.WriteRune(r)
				case ' ':
					if ct == 0 {
						continue
					}
					if ct < 100 {
						buf.WriteRune(r)
						continue
					}
					ct = 0
					buf.WriteString("\n// ")
				case '\n':
					ct = 0
					buf.WriteString("\n// ")
				}
			}

			return buf.String()
		},
		"typeprefix": func(f Field) string {
			if f.Rpt != "1" || f.Rpt == "*" {
				return "[]"
			}
			// Special case these, where empty is not present.
			switch f.DataType {
			default:
				if emptyDataTypes[f.DataType] {
					return ""
				}
			case "TS", "TM", "DTM", "DT":
				return ""
			case "VARIES":
				// continue.
			case "FN":
				return ""
			}
			switch f.Usage {
			case "O", "C":
				return "*"
			}
			return ""
		},
		"tag": func(f Field) string {
			buf := &strings.Builder{}
			_, pos, _ := strings.Cut(f.Position, ".")
			buf.WriteString(pos)

			switch f.ID {
			case "FieldSeparator":
				buf.WriteString(",noescape,fieldsep,omit")
			case "EncodingCharacters":
				buf.WriteString(",noescape,fieldchars")
			case "SetID":
				buf.WriteString(",seq")
			}
			switch f.Usage {
			case "R":
				buf.WriteString(",required")
			case "C":
				buf.WriteString(",conditional")
			}
			if f.Rpt != "1" && f.Rpt != "*" {
				buf.WriteString(",max=")
				buf.WriteString(displaySan.Replace(f.Rpt))
			}
			if f.Length > 0 {
				buf.WriteString(",len=")
				buf.WriteString(strconv.FormatInt(int64(f.Length), 10))
			}
			if len(f.TableID) > 0 {
				buf.WriteString(",table=")
				buf.WriteString(displaySan.Replace(f.TableID))
			}
			switch f.DataType {
			case "TM":
				buf.WriteString(",format=HM")
			case "DT":
				buf.WriteString(",format=YMD")
			case "DTM":
				buf.WriteString(",format=YMDHM")
			case "TS":
				buf.WriteString(",format=YMDHMS")
			}
			desc := f.Description
			if len(desc) == 0 {
				desc = f.Name
			}
			if len(desc) > 0 {
				buf.WriteString(",display=")
				buf.WriteString(displaySan.Replace(desc))
			}

			return buf.String()
		},
		"seqtypeprefix": func(f TriggerSegment) string {
			if f.Rpt != "1" || f.Rpt == "*" {
				return "[]"
			}
			switch f.Usage {
			case "O", "C":
				return "*"
			}
			// This enabled the unmarshal to work correctly on deep groups currently.
			return "*"
		},
		"segtag": func(f TriggerSegment) string {
			buf := &strings.Builder{}
			buf.WriteString(f.Sequence)

			switch f.Usage {
			case "R":
				buf.WriteString(",required")
			case "C":
				buf.WriteString(",conditional")
			}
			if f.Rpt != "1" && f.Rpt != "*" {
				buf.WriteString(",max=")
				buf.WriteString(displaySan.Replace(f.Rpt))
			}
			desc := f.LongName
			if len(desc) == 0 {
				desc = f.Name
			}
			if len(desc) > 0 {
				buf.WriteString(",display=")
				buf.WriteString(displaySan.Replace(desc))
			}

			return buf.String()
		},
		"unique": func(list []TableRow) []TableRow {
			ret := make([]TableRow, 0, len(list))
			seen := map[string]bool{}
			for _, row := range list {
				if seen[row.Value] {
					continue
				}
				seen[row.Value] = true
				ret = append(ret, row)
			}
			return ret
		},
		"pack": func(vv ...any) map[string]any {
			ret := make(map[string]any, len(vv)/2)
			var key string
			for _, v := range vv {
				if len(key) == 0 {
					key = v.(string)
					continue
				}
				ret[key] = v
				key = ""
			}
			return ret
		},
		"memorize": func(id, name string, list []TriggerSegment) string {
			key := memorizeKey{
				id:   id,
				name: name,
			}
			if memorizeUnique[key] {
				// panic(fmt.Errorf("seen %v", key))
			}
			memorizeUnique[key] = true
			memorizeList = append(memorizeList, memorizeValue{
				ID:       id,
				Name:     name,
				Segments: list,
			})
			return ""
		},
		"pop": func() []memorizeValue {
			list := memorizeList
			memorizeList = memorizeList[:0]
			return list
		},
		"messageType": func(st SegmentType) string {
			for _, f := range st.Fields {
				switch f.DataType {
				case "CM_MSG", "MSG":
					return f.ID
				}
			}
			return ""
		},
		"hasField": func(name string, ff []Field) bool {
			for _, f := range ff {
				if f.ID == name {
					return true
				}
			}
			return false
		},
	}

	temp, err := template.New("").Funcs(funcs).ParseFS(pkgTemplate, "template/*.go.template")
	if err != nil {
		return err
	}
	err = os.MkdirAll(r.pkgDir, 0700)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	for _, temp := range temp.Templates() {
		if !strings.HasSuffix(temp.Name(), ".template") {
			continue
		}
		name := strings.TrimSuffix(temp.Name(), ".template")
		err = temp.Execute(buf, to)
		if err != nil {
			return fmt.Errorf("template %s: %w", name, err)
		}
		fn := filepath.Join(r.pkgDir, name)
		bfmt, err := format.Source(buf.Bytes())
		if err != nil {
			os.WriteFile(fn, buf.Bytes(), 0600)
			return fmt.Errorf("format %s: %w", name, err)
		}
		err = os.WriteFile(fn, bfmt, 0600)
		if err != nil {
			return err
		}
		buf.Reset()
	}

	return nil
}

func visitAllDataType(r *runner, seen map[string]bool, dtName string, visit func(dt DataType) error) error {
	if len(dtName) == 0 {
		return nil
	}
	dt, err := r.getDataType(dtName)
	if err != nil {
		return err
	}
	seen[dt.Name] = true

	err = visit(dt)
	if err != nil {
		return err
	}

	for _, f := range dt.Fields {
		if len(f.DataType) == 0 {
			continue
		}
		if seen[f.DataType] {
			continue
		}
		err = visitAllDataType(r, seen, f.DataType, visit)
		if err != nil {
			return err
		}
	}
	return nil
}

func visitAllSegment(top string, parent *TriggerSegment, list []TriggerSegment, f func(top string, parent, segment *TriggerSegment) error) error {
	var err error
	for i := range list {
		s := &list[i]
		err = f(top, parent, s)
		if err != nil {
			return err
		}
		if len(s.Segments) > 0 {
			err = visitAllSegment(top, s, s.Segments, f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
