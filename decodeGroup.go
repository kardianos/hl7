package hl7

import (
	"errors"
	"fmt"
	"reflect"
)

type messageStructure interface {
	MessageStructureID() []string
}

func newWalker(list []any, registry Registry) (*walker, error) {
	if len(list) == 0 {
		return nil, fmt.Errorf("list is empty")
	}

	root := list[0]
	ms, ok := root.(messageStructure)
	if !ok {
		return nil, ErrUnexpectedSegment{
			Trigger:    fmt.Sprintf("first message must implment MessageStructure, %T does not", root),
			LineNumber: 1,
			Segment:    root,
		}
	}
	codeList := ms.MessageStructureID()
	if len(codeList) == 0 {
		return nil, fmt.Errorf("message structure code missing, malformed message: %#v", root)
	}
	var vex any
	var code string
	for _, c := range codeList {
		if len(c) == 0 {
			return nil, fmt.Errorf("message structure code empty, malformed message: %#v", root)
		}
		vex, ok = registry.Trigger(c)
		if ok {
			code = c
			break
		}
	}
	if vex == nil {
		return nil, fmt.Errorf("message structure code not found %q", codeList)
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
	return rv.IsValid() && !rv.IsZero()
}

type structItem struct {
	Index       int
	Parent      *structItem
	LinkType    linkType
	Type        reflect.Type
	ActiveValue reflect.Value
	Leaf        bool // Is a leaf node, nothing beyond this node.
	InArray     bool // If this struct node is within an array (and thus can always be added to).
	Depth       int  // Depth in the tree (0 = root, higher = deeper).
}

func (si *structItem) present() bool {
	return present(si.ActiveValue)
}

// currentValue returns the actual current value for this structItem by traversing
// the parent chain. This is necessary because ActiveValue can become stale when
// new list entries are created - the parent's ActiveValue is updated but the
// children's ActiveValue still points to the old values.
func (si *structItem) currentValue() reflect.Value {
	if si.Parent == nil {
		return si.ActiveValue
	}
	// For list items, ActiveValue points to the current list element.
	// Return it directly because for lists, the "current value" is the
	// current element we're working on, not the slice from the parent.
	if si.LinkType == linkList {
		return si.ActiveValue
	}

	// Get the parent's current value.
	// For list items (linkList), the parent's ActiveValue points to the current
	// list element, so we use that directly. For non-list items, we recursively
	// traverse up the parent chain.
	var parentValue reflect.Value
	switch si.Parent.LinkType {
	case linkList:
		parentValue = si.Parent.ActiveValue
	default:
		parentValue = si.Parent.currentValue()
	}
	if !parentValue.IsValid() {
		return reflect.Value{}
	}

	// Dereference pointers.
	for {
		switch parentValue.Kind() {
		case reflect.Pointer:
			if parentValue.IsNil() {
				return reflect.Value{}
			}
			parentValue = parentValue.Elem()
			continue
		case reflect.Struct:
			return parentValue.Field(si.Index)
		}
		return reflect.Value{}
	}
}

// presentInContext checks if this structItem has a value set in the current
// context. Unlike present(), this handles the case where parent list entries
// have been created and the ActiveValue is stale.
func (si *structItem) presentInContext() bool {
	return present(si.currentValue())
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

	// Find all valid candidates (both forward and backward).
	var candidates []*candidateMatch

	// First look forward.
	for i := w.last; i < len(w.list); i++ {
		si := w.list[i]
		if si.Type != rt {
			continue
		}
		if w.fullInArray(si) {
			continue
		}
		candidates = append(candidates, &candidateMatch{index: i, si: si, forward: true})
	}
	// Then look backwards.
	for i := w.last - 1; i >= 0; i-- {
		si := w.list[i]
		if si.Type != rt {
			continue
		}
		if w.fullInArray(si) {
			continue
		}
		candidates = append(candidates, &candidateMatch{index: i, si: si, forward: false})
	}

	if len(candidates) > 0 {
		// Select the best candidate based on these priorities:
		// 1. Find the shallowest forward match and shallowest backward match.
		// 2. If a backward match exists and is shallower than ALL forward matches,
		//    use the backward match. This handles starting new repeating groups
		//    (e.g., new Specimen) when a segment type appears at multiple depths.
		// 3. Otherwise, use the shallowest forward match to maintain tree order.
		var forwardCandidates, backwardCandidates []*candidateMatch
		for _, c := range candidates {
			if c.forward {
				forwardCandidates = append(forwardCandidates, c)
			} else {
				backwardCandidates = append(backwardCandidates, c)
			}
		}

		var best *candidateMatch

		// Find shallowest forward match
		var shallowestForward *candidateMatch
		if len(forwardCandidates) > 0 {
			shallowestForward = forwardCandidates[0]
			for _, c := range forwardCandidates[1:] {
				if c.si.Depth < shallowestForward.si.Depth {
					shallowestForward = c
				}
			}
		}

		// Find shallowest backward match
		var shallowestBackward *candidateMatch
		if len(backwardCandidates) > 0 {
			shallowestBackward = backwardCandidates[0]
			for _, c := range backwardCandidates[1:] {
				if c.si.Depth < shallowestBackward.si.Depth {
					shallowestBackward = c
				}
			}
		}

		// Get the depth of the current position to inform the decision.
		var currentDepth int
		if w.last >= 0 && w.last < len(w.list) {
			currentDepth = w.list[w.last].Depth
		}

		// Decision logic:
		// - Default: prefer forward to maintain natural segment order.
		// - Only prefer backward when we need to "break out" to start a new repeating group.
		//   This requires: (1) backward is in an array context (InArray=true), meaning it can
		//   start a new group, and (2) there's a significant depth difference (>= 2) indicating
		//   we're deep in a structure and need to go back to a shallower level.
		// - Otherwise, use forward (or backward if no forward exists).
		backwardStartsNewGroup := shallowestBackward != nil &&
			shallowestBackward.si.InArray &&
			currentDepth-shallowestBackward.si.Depth >= 2

		switch {
		case backwardStartsNewGroup && (shallowestForward == nil || shallowestBackward.si.Depth < shallowestForward.si.Depth):
			best = shallowestBackward
		case shallowestForward != nil:
			best = shallowestForward
		default:
			best = shallowestBackward
		}

		return w.found(best.index, best.si, rv)
	}

	// Control segments are handled separately.
	if _, isControl := w.registry.ControlSegment(rt.Name()); isControl {
		// TODO: handle batch and control segments.
		return nil
	}
	return ErrUnexpectedSegment{
		Trigger:    w.triggerCode,
		LineNumber: line,
		Segment:    v,
	}
}

type candidateMatch struct {
	index   int
	si      *structItem
	forward bool
}

// The error ErrUnexpectedSegment will be returned if an unexpected segment for a
// given trigger is read in the message.
//
//	var segment hl7.ErrUnexpectedSegment
//	if errors.As(err, &segment) { /* Use segment. */ }
type ErrUnexpectedSegment struct {
	Trigger    string // Name of the trigger.
	LineNumber int    // Line number of the message this segment is found on.
	Segment    any    // Segment value, such as *h250.MSA.
}

func (err ErrUnexpectedSegment) Error() string {
	return fmt.Sprintf("line %d (%T) not found in trigger %q", err.LineNumber, err.Segment, err.Trigger)
}

// found creates the parent tree and sets it up.
func (w *walker) found(index int, si *structItem, rv reflect.Value) error {
	w.last = index

	currentList := []*structItem{}
	current := si
	findList := si.presentInContext() // Current value is full, find the next list.
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
		valid := current.presentInContext()
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
		if c.Leaf {
			set = rv
		} else {
			set = reflect.New(c.Type)
		}

		pv := c.Parent.ActiveValue.Field(c.Index)
		setKind := set.Kind()

		switch c.LinkType {
		case linkValue:
			if setKind == reflect.Pointer {
				set = set.Elem()
			}
			if present(pv) {
				return fmt.Errorf("expected empty value %s, value present", pv.Type())
			}
			pv.Set(set)
			c.set(pv)

		case linkOpt:
			if setKind != reflect.Pointer {
				set = set.Addr()
			}
			if present(pv) {
				return fmt.Errorf("expected empty pointer %s, pointer is present", pv.Type())
			}
			pv.Set(set)
			c.set(pv.Elem())

		case linkList:
			if setKind == reflect.Pointer {
				set = set.Elem()
			}
			pv.Set(reflect.Append(pv, set))
			c.set(pv.Index(pv.Len() - 1))
		}
	}

	return nil
}

// fullInArray returns true if there is no place to put another value by setting
// an existing value or by adding an item (either self or parent).
func (w *walker) fullInArray(si *structItem) bool {
	if si.InArray {
		return false
	}
	av := si.ActiveValue
	if !av.IsValid() {
		return false
	}
	// Structs are never "full" in this context; nil/zero values are not full.
	switch av.Kind() {
	case reflect.Struct:
		return false
	default:
		return !av.IsNil() && !av.IsZero()
	}
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

	depth := 0
	if parent != nil {
		depth = parent.Depth + 1
	}
	item := &structItem{
		Index:    fieldIndex,
		Parent:   parent,
		LinkType: link,
		Type:     baseType,
		Leaf:     leaf,
		InArray:  inArray,
		Depth:    depth,
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
