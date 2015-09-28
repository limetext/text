// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"fmt"
)

type (
	// Action defines an interface for Apply-ing and Undo-ing
	// an action.
	Action interface {
		Apply()
		Undo()
	}

	// CompositeAction is just a structure containing multiple Action items
	CompositeAction struct {
		actions []Action
	}

	insertAction struct {
		buffer Buffer
		point  int
		value  []rune
	}

	eraseAction struct {
		insertAction
		region Region
	}
)

func (ca CompositeAction) String() string {
	ret := fmt.Sprintf("%d actions:\n", len(ca.actions))
	for i := range ca.actions {
		ret += fmt.Sprintf("\t%s\n", ca.actions[i])
	}
	return ret
}

// Apply all the sub-actions in order of this CompositeAction
func (ca *CompositeAction) Apply() {
	for _, a := range ca.actions {
		a.Apply()
	}
}

// Undo all the sub-actions in reverse order of this CompositeAction
func (ca *CompositeAction) Undo() {
	l := len(ca.actions) - 1
	for i := range ca.actions {
		ca.actions[l-i].Undo()
	}
}

// Add adds the action to this CompositeAction, without first
// executing the action
func (ca *CompositeAction) Add(a Action) {
	ca.actions = append(ca.actions, a)
}

// AddExec executes the provided action and then adds
// the action to this CompositeAction
func (ca *CompositeAction) AddExec(a Action) {
	ca.Add(a)
	ca.actions[len(ca.actions)-1].Apply()
}

// Len returns the number of sub-actions this CompositeAction object contains
func (ca *CompositeAction) Len() int {
	return len(ca.actions)
}

func (ia *insertAction) Apply() {
	ia.buffer.Insert(ia.point, string(ia.value))
}

func (ia *insertAction) Undo() {
	ia.buffer.Erase(ia.point, len(ia.value))
}

func (ea *eraseAction) Apply() {
	ea.region = ea.region.Intersection(Region{0, ea.buffer.Size()})
	ea.value = []rune(ea.buffer.Substr(ea.region))
	ea.point = ea.region.Begin()
	ea.insertAction.Undo()
}

func (ea *eraseAction) Undo() {
	ea.insertAction.Apply()
}

func (ia insertAction) String() string {
	return fmt.Sprintf("insert %d %s", ia.point, string(ia.value))
}

func (ea eraseAction) String() string {
	return fmt.Sprintf("erase %v", ea.region)
}

// NewEraseAction returns a new action that erases the given region in the given buffer
func NewEraseAction(b Buffer, region Region) Action {
	return &eraseAction{insertAction{buffer: b}, region}
}

// NewInsertAction returns a new action that inserts the given string value at
// position point in the Buffer b.
func NewInsertAction(b Buffer, point int, value string) Action {
	return &insertAction{b, Clamp(0, b.Size(), point), []rune(value)}
}

// NewReplaceAction returns a new action that replaces the data in the given
// buffers region with the provided value
func NewReplaceAction(b Buffer, region Region, value string) Action {
	return &CompositeAction{[]Action{
		NewEraseAction(b, region),
		NewInsertAction(b, Clamp(0, b.Size()-region.Size(), region.Begin()), value),
	}}
}
