package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var errIgnore = fmt.Errorf("ignore")

type runner struct {
	version      string
	rootDir      string
	allowNetwork bool
	pkgDir       string

	cache  map[cacheKey]any
	client *http.Client
}

type cacheKey struct {
	Version  string
	Resource resource
	Name     string
}

func (r *runner) urlPath(res resource, name string) string {
	switch {
	default:
		return rootURL + r.version + "/" + string(res) + "/" + name
	case len(name) == 0:
		return rootURL + r.version + "/" + string(res)
	}
}

func (r *runner) filePath(res resource, name string, old bool) string {
	// Do not use reserved names on Windows.
	if !old {
		switch name {
		case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
			name = name + "_X"
		}
	}
	switch {
	default:
		return filepath.Join(r.rootDir, r.version, string(res), name+".json")
	case len(name) == 0:
		return filepath.Join(r.rootDir, r.version, string(res), "list.json")
	}
}

func (r *runner) getJSON(res resource, name string) ([]byte, error) {
	fn := r.filePath(res, name, false)
	bb, err := os.ReadFile(fn)
	if os.IsNotExist(err) {
		fnOld := r.filePath(res, name, true)
		bb, err = os.ReadFile(fnOld)
		if err == nil {
			os.Rename(fnOld, fn)
		}
	}
	if os.IsNotExist(err) {
		const ignoreSuffix = ".ignore"
		if _, err := os.Stat(fn + ignoreSuffix); err == nil {
			return nil, errIgnore
		}
		if !r.allowNetwork {
			return nil, fmt.Errorf("missing %s, network not allowed", fn)
		}

		dir, _ := filepath.Split(fn)
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return nil, err
		}
		u := r.urlPath(res, name)

		bb, err = r.fetch(u)
		if err == errIgnore {
			err = os.WriteFile(fn+ignoreSuffix, []byte("# Ignore. Not present in version."), 0600)
			if err != nil {
				return nil, err
			}
			return nil, errIgnore
		}
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(fn, bb, 0600)
		if err != nil {
			return nil, err
		}
		return bb, nil
	}
	if err != nil {
		return nil, err
	}
	return bb, nil
}

func (r *runner) log(v ...any) {
	log.Print(v...)
}

func (r *runner) fetch(u string) ([]byte, error) {
	r.log("fetch", u)
	resp, err := r.client.Get(u)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	default:
		return nil, fmt.Errorf("status %s:\n%s", resp.Status, buf.String())
	case 404:
		return nil, errIgnore
	case 200:
		return buf.Bytes(), nil
	}
}

func get[T any](r *runner, res resource, name string) (T, error) {
	ck := cacheKey{
		Version:  r.version,
		Resource: res,
		Name:     name,
	}
	ci, ok := r.cache[ck]
	if ok {
		return ci.(T), nil
	}

	v := new(T)
	bb, err := r.getJSON(res, name)
	if err != nil {
		return *v, fmt.Errorf("for %s/%s: %w", res, name, err)
	}
	err = json.Unmarshal(bb, v)
	if err != nil {
		return *v, fmt.Errorf("for %s/%s: %w\n\n%s", res, name, err, bb)
	}
	r.cache[ck] = *v
	return *v, nil
}

func (r *runner) getChapterList() ([]Chapter, error) {
	return get[[]Chapter](r, ResourceChapters, "")
}

func (r *runner) getChapter(name string) (Chapter, error) {
	return get[Chapter](r, ResourceChapters, name)
}
func (r *runner) getTriggerList() ([]TriggerEvents, error) {
	return get[[]TriggerEvents](r, ResourceTriggerEvents, "")
}
func (r *runner) getTrigger(name string) (TriggerType, error) {
	return get[TriggerType](r, ResourceTriggerEvents, name)
}
func (r *runner) getSegment(name string) (SegmentType, error) {
	return get[SegmentType](r, ResourceSegments, name)
}
func (r *runner) getDataType(name string) (DataType, error) {
	return get[DataType](r, ResourceDataTypes, name)
}
func (r *runner) getTable(name string) (TableType, error) {
	return get[TableType](r, ResourceTables, name)
}
