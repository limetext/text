// Copyright 2014 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"reflect"
	"testing"
)

type (
	observerData struct {
		buf Buffer
		r   Region
		d   []rune
	}
	dummyObserver struct {
		badlyBehaved bool
		inserted     observerData
		erased       observerData
	}
)

func (do *dummyObserver) Inserted(b Buffer, r Region, d []rune) {
	if do.badlyBehaved {
		if err := b.Erase(0, 1); err == nil {
			panic("Expected an error here")
		}
		if err := b.Insert(0, "test"); err == nil {
			panic("Expected an error here")
		}
		if err := b.InsertR(0, []rune("test")); err == nil {
			panic("Expected an error here")
		}
	}
	do.inserted = observerData{b, r, d}
}
func (do *dummyObserver) Erased(b Buffer, r Region, d []rune) {
	if do.badlyBehaved {
		if err := b.Erase(0, 1); err == nil {
			panic("Expected an error here")
		}
		if err := b.Insert(0, "test"); err == nil {
			panic("Expected an error here")
		}
		if err := b.InsertR(0, []rune("test")); err == nil {
			panic("Expected an error here")
		}
	}
	do.erased = observerData{b, r, d}
}

func TestBufferObserver(t *testing.T) {
	var b = NewBuffer()
	defer b.Close()

	do1 := &dummyObserver{}
	do2 := &dummyObserver{}

	// Adding
	if err := b.AddObserver(do1); err != nil {
		t.Fatal(err)
	}
	if err := b.AddObserver(do2); err != nil {
		t.Fatal(err)
	}

	// Re-adding
	if err := b.AddObserver(do1); err == nil {
		t.Fatal("Expected an error but didn't get one!")
	}
	if err := b.AddObserver(do2); err == nil {
		t.Fatal("Expected an error but didn't get one!")
	}

	// Removing
	if err := b.RemoveObserver(do1); err != nil {
		t.Fatal(err)
	}
	// Re-removing
	if err := b.RemoveObserver(do1); err == nil {
		t.Fatal("Expected an error but didn't get one!")
	}

	// Ok, now that we've got that over with, lets try the actual
	// callbacks
	if err := b.AddObserver(do1); err != nil {
		t.Fatal(err)
	}

	data := []rune("hello world")
	exp := observerData{b, Region{0, len(data)}, data}
	b.InsertR(exp.r.Begin(), exp.d)
	if !reflect.DeepEqual(do1, do2) {
		t.Errorf("do1's and do2's data aren't equal:\n%+v\n%+v", do1, do2)
	} else if !reflect.DeepEqual(do1.inserted, exp) {
		t.Errorf("do1's and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	}

	data = []rune("llo w")
	exp = observerData{b, Region{2, 2 + len(data)}, data}
	b.Erase(exp.r.Begin(), exp.r.Size())
	if !reflect.DeepEqual(do1, do2) {
		t.Errorf("do1's and do2's data aren't equal:\n%+v\n%+v", do1, do2)
	} else if !reflect.DeepEqual(do1.erased, exp) {
		t.Errorf("do1's and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	} else if sub := b.Substr(Region{0, b.Size()}); sub != "heorld" {
		t.Errorf("Unexpected buffer contents: %s", sub)
	}

	// Just to make sure all the badly behaved operations are all nops
	do1.inserted.buf = nil
	do1.badlyBehaved = true
	b.Insert(exp.r.Begin(), string(exp.d))
	do1.badlyBehaved = false
	if !reflect.DeepEqual(do1, do2) {
		t.Errorf("do1's and do2's data aren't equal:\n%+v\n%+v", do1, do2)
	} else if !reflect.DeepEqual(do1.inserted, exp) {
		t.Errorf("do1's inserted and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	} else if !reflect.DeepEqual(do1.erased, exp) {
		t.Errorf("do1's erased and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	} else if sub := b.Substr(Region{0, b.Size()}); sub != "hello world" {
		t.Errorf("Unexpected buffer contents: %s", sub)
	}

	do1.erased.buf = nil
	do1.badlyBehaved = true
	b.Erase(exp.r.Begin(), exp.r.Size())
	do1.badlyBehaved = false
	if !reflect.DeepEqual(do1, do2) {
		t.Errorf("do1's and do2's data aren't equal:\n%+v\n%+v", do1, do2)
	} else if !reflect.DeepEqual(do1.inserted, exp) {
		t.Errorf("do1's inserted and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	} else if !reflect.DeepEqual(do1.erased, exp) {
		t.Errorf("do1's erased and exp's data aren't equal:\n%+v\n%+v", do1.inserted, exp)
	} else if sub := b.Substr(Region{0, b.Size()}); sub != "heorld" {
		t.Errorf("Unexpected buffer contents: %s", sub)
	}

}
