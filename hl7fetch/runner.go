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

type Runner struct {
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

func (r *Runner) urlPath(res resource, name string) string {
	switch {
	default:
		return rootURL + r.version + "/" + string(res) + "/" + name
	case len(name) == 0:
		return rootURL + r.version + "/" + string(res)
	}
}

func (r *Runner) filePath(res resource, name string) string {
	switch {
	default:
		return filepath.Join(r.rootDir, r.version, string(res), name+".json")
	case len(name) == 0:
		return filepath.Join(r.rootDir, r.version, string(res), "list.json")
	}
}

func (r *Runner) getJSON(res resource, name string) ([]byte, error) {
	fn := r.filePath(res, name)
	bb, err := os.ReadFile(fn)
	if os.IsNotExist(err) {
		if !r.allowNetwork {
			return nil, fmt.Errorf("missing %s, network not allowed", fn)
		}
		u := r.urlPath(res, name)
		bb, err = r.fetch(u)
		if err != nil {
			return nil, err
		}
		dir, _ := filepath.Split(fn)
		err = os.MkdirAll(dir, 0700)
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

func (r *Runner) log(v ...any) {
	log.Print(v...)
}

func (r *Runner) fetch(u string) ([]byte, error) {
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
	case 200:
		return buf.Bytes(), nil
	}
}

func get[T any](r *Runner, res resource, name string) (T, error) {
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

func (r *Runner) getChapterList() ([]Chapter, error) {
	return get[[]Chapter](r, ResourceChapters, "")
}

func (r *Runner) getChapter(name string) (Chapter, error) {
	return get[Chapter](r, ResourceChapters, name)
}
func (r *Runner) getTriggerList() ([]TriggerEvents, error) {
	return get[[]TriggerEvents](r, ResourceTriggerEvents, "")
}
func (r *Runner) getTrigger(name string) (TriggerType, error) {
	return get[TriggerType](r, ResourceTriggerEvents, name)
}
func (r *Runner) getSegment(name string) (SegmentType, error) {
	return get[SegmentType](r, ResourceSegments, name)
}
func (r *Runner) getDataType(name string) (DataType, error) {
	return get[DataType](r, ResourceDataTypes, name)
}
func (r *Runner) getTable(name string) (TableType, error) {
	return get[TableType](r, ResourceTables, name)
}
