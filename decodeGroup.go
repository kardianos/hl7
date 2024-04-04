package hl7

import (
	"errors"
	"fmt"
	"reflect"
)

type messageStructure interface {
	MessageStructureID() string
}

func newWalker(list []any, registry Registry) (*walker, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("list is empty")
	}

	root := list[0]
	ms, ok := root.(messageStructure)
	if !ok {
		return nil, fmt.Errorf("first message must implment MessageStructure, %T does not", root)
	}
	code := ms.MessageStructureID()
	if len(code) == 0 {
		return nil, fmt.Errorf("message structure code empty, malformed message: %#v", root)
	}
	vex, ok := registry.Trigger(code)
	if !ok {
		return nil, fmt.Errorf("message structure code not found %q", code)
	}
	tp := reflect.TypeOf(vex)

	// Map a linear structure onto a hierarchical structure.
	//
	// 1. Go through trigger struct, depth first.
	// 2. Create a tree, noting the single, optional, and repeated fields, as well as sequence numbers.
	// 3. Loop A: Starting at sequence zero, go forward to match.
	// 4. If no match from current location, go backwards to find match.
	// 5. If nothing matches, return error.
	// 6. Determine if array item(s) needs to be added in parent tree.
	// 7. Add any parent items.
	// 8. Add segment.
	// 9. Go to next segment. End Loop A.
	// 10. When all segments are processed, return trigger structure.

	w := &walker{
		triggerCode: code,
		registry:    registry,
	}
	err := w.eat(nil, 0, tp, false)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (w *walker) process(list []any) (any, error) {
	for i, item := range list {
		err := w.digest(i+1, item)
		if err != nil {
			return nil, err
		}
	}

	if len(w.list) == 0 {
		return nil, fmt.Errorf("missing list item,his should not happen")
	}
	rootSI := w.list[0]
	if !rootSI.ActiveValue.IsValid() {
		return nil, fmt.Errorf("root value nil, input %q", list)
	}
	rootI := rootSI.ActiveValue.Interface()

	return rootI, nil
}

// group a list of segments into hierarchical groups with a single root element.
func group(list []any, registry Registry) (any, error) {
	var segErrs []error
	for i, item := range list {
		if se, ok := item.(SegmentError); ok {
			segErrs = append(segErrs, se.ErrorList...)
			list[i] = se.Segment
		}
	}
	w, err := newWalker(list, registry)
	if err != nil {
		return nil, err
	}
	gr, err := w.process(list)
	if err != nil {
		segErrs = append(segErrs, err)
	}
	return gr, errors.Join(segErrs...)
}

type linkType int

const (
	linkUnknown linkType = iota
	linkValue
	linkOpt
	linkList
)

func (lt linkType) String() string {
	switch lt {
	default:
		return ""
	case linkUnknown:
		return "unknown"
	case linkValue:
		return "value"
	case linkOpt:
		return "opt"
	case linkList:
		return "list"
	}
}

func present(rv reflect.Value) bool {
	if !rv.IsValid() {
		return false
	}
	if rv.IsZero() {
		return false
	}

	return true
}

type structItem struct {
	Index       int
	Parent      *structItem
	LinkType    linkType
	Type        reflect.Type
	ActiveValue reflect.Value
	Leaf        bool // Is a leaf node, nothing beyond this node.
	InArray     bool // If this struct node is within an array (and thus can always be added to).
}

func (si *structItem) present() bool {
	return present(si.ActiveValue)
}

func (si *structItem) set(rv reflect.Value) {
	si.ActiveValue = rv
}

type walker struct {
	triggerCode string // For error reporting.
	registry    Registry

	last int
	list []*structItem
}

func (w *walker) digest(line int, v any) error {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	if rt.Kind() == reflect.Pointer {
		rt = rt.Elem()
	}
	// First look forward.
	for i := w.last; i < len(w.list); i++ {
		si := w.list[i]
		if si.Type != rt {
			continue
		}
		if w.fullInArray(si) {
			continue
		}
		return w.found(line, i, si, v, rv, rt)
	}
	// If not found going forward, go backwards.
	for i := w.last; i >= 0; i-- {
		si := w.list[i]
		if si.Type != rt {
			continue
		}
		if w.fullInArray(si) {
			continue
		}
		return w.found(line, i, si, v, rv, rt)
	}

	_, isControl := w.registry.ControlSegment(rt.Name())
	if isControl {
		// TODO: handle batch and control segments.
		return nil
	}
	return fmt.Errorf("line %d (%T) not found in trigger %q", line, v, w.triggerCode)
}

// Found creates the parent tree and sets it up.
func (w *walker) found(line, index int, si *structItem, v any, rv reflect.Value, rt reflect.Type) error {
	w.last = index

	// fmt.Printf("found, %s\n", rv.Type())
	currentList := []*structItem{}
	current := si
	findList := si.present() // Current value is full, find the next list.
	hasList := false
	for {
		// Stop at the root item or at the first parent which is valid.
		if current == nil {
			break
		}
		// Continue stepping down until either the Active value is not empty or a slice is encountered.
		// If the current item is not valid, always add.
		// If the current item is valid, but full, continue.
		// A LinkList LinkType is never full.
		// valid := present(current.ActiveValue)
		valid := current.present()
		if !valid {
			if current.LinkType == linkList {
				hasList = true
			}
			currentList = append(currentList, current)
			current = current.Parent
			continue
		}
		// All valid.
		// Break if no list is needed or if a list has already been found.
		if !findList || hasList {
			break
		}
		currentList = append(currentList, current)
		if current.LinkType == linkList {
			break
		}
		current = current.Parent
	}
	if len(currentList) == 0 {
		return fmt.Errorf("found, nothing in current list")
	}

	for i := len(currentList) - 1; i >= 0; i-- {
		c := currentList[i]
		if c.Parent == nil {
			// The root value just needs to be created when found.
			if c.present() {
				parent := c.ActiveValue.Type().String()
				child := rv.Type().String()
				return fmt.Errorf("cannot overwrite %[2]s in %[1]s when %[2]s is already present", parent, child)
			}
			c.ActiveValue = reflect.New(c.Type).Elem()
			continue
		}
		var set reflect.Value
		switch {
		case c.Leaf:
			set = rv
		default:
			set = reflect.New(c.Type)
		}
		pv := c.Parent.ActiveValue.Field(c.Index)
		switch c.LinkType {
		case linkValue:
			if set.Kind() == reflect.Pointer {
				set = set.Elem()
			}
			if present(pv) {
				return fmt.Errorf("expected empty value %s, value present", pv.Type())
			}
			pv.Set(set)
			c.set(pv)

		case linkOpt:
			if set.Kind() != reflect.Pointer {
				set = set.Addr()
			}
			if present(pv) {
				return fmt.Errorf("expected empty pointer %s, pointer is present", pv.Type())
			}
			pv.Set(set)
			c.set(pv.Elem())

		case linkList:
			if set.Kind() == reflect.Pointer {
				set = set.Elem()
			}
			pv.Set(reflect.Append(pv, set))
			c.set(pv.Index(pv.Len() - 1))
		}
	}

	return nil
}

// Returns true if there is no place to put another value by setting an existing value
// or by adding an item (either selv or parent).
func (w *walker) fullInArray(si *structItem) bool {
	if si.InArray {
		return false
	}
	// If not valid, if pointer is nil, or if zero value, then not full.
	if !si.ActiveValue.IsValid() {
		return false
	}
	if kind := si.ActiveValue.Kind(); kind == reflect.Struct {
		return false
	}
	if si.ActiveValue.IsNil() {
		return false
	}
	if si.ActiveValue.IsZero() {
		return false
	}
	return true
}

func (w *walker) eat(parent *structItem, fieldIndex int, rt reflect.Type, inArray bool) error {
	var baseType reflect.Type
	var link linkType
	switch rt.Kind() {
	default:
		return fmt.Errorf("unknown kind: %v", rt.Kind())
	case reflect.Struct:
		link = linkValue
		baseType = rt
	case reflect.Slice:
		link = linkList
		baseType = rt.Elem()
		inArray = true
	case reflect.Pointer:
		link = linkOpt
		baseType = rt.Elem()
	}
	var currentTag tag
	if metaField, ok := baseType.FieldByName(hl7MetaName); ok {
		var err error
		currentTag, err = parseTag(metaField.Name, metaField.Tag.Get(tagName))
		if err != nil {
			return err
		}
	}
	leaf := parent != nil
	switch currentTag.Type {
	case structTriggerGroup:
		leaf = false
	}

	item := &structItem{
		Index:    fieldIndex,
		Parent:   parent,
		LinkType: link,
		Type:     baseType,
		Leaf:     leaf,
		InArray:  inArray,
	}
	w.list = append(w.list, item)

	if leaf {
		return nil
	}

	// Peek into linked struct type, get meta info.

	// For each field type we look at, be sure to look at the tag type.
	// Only look at "t" and "tg" types. The segments must be the leaf types.

	ct := baseType.NumField()
	for i := 0; i < ct; i++ {
		ft := baseType.Field(i)
		tag, err := parseTag(ft.Name, ft.Tag.Get(tagName))
		if err != nil {
			return err
		}
		if !tag.Present {
			continue
		}
		if tag.Meta {
			continue
		}
		err = w.eat(item, i, ft.Type, inArray)
		if err != nil {
			return err
		}
	}

	return nil
}
