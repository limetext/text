// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"runtime"
	"strings"
	"sync"
)

type (
	// The InnerBufferInterface defines a minimal
	// interface that different buffer implementations
	// need to implement.
	InnerBufferInterface interface {
		// Returns the size of a buffer
		Size() int

		// Returns the runes in the specified Region.
		// Implementations should clamp the region
		// as appropriate.
		SubstrR(r Region) []rune

		// Inserts the given rune data at the
		// specified point in the buffer.
		InsertR(point int, data []rune)

		// Erases "length" units from "point".
		Erase(point, length int)

		// Returns the rune at the given index
		Index(int) rune

		// Convert a text position into a row and column.
		RowCol(point int) (row, col int)

		// Inverse of #RowCol, converting a row and column
		// into a text position.
		TextPoint(row, col int) (i int)

		// Close the buffer, freeing any associated resources.
		Close()
	}

	// The full buffer interface defines
	Buffer interface {
		InnerBufferInterface
		IdInterface
		SettingsInterface
		sync.Locker

		// Add a BufferChangedCallback to the buffer
		AddCallback(cb BufferChangedCallback)

		SetName(string)
		Name() string
		SetFileName(string)
		FileName() string

		// Inserts the given string at the given location.
		// Typically just a wrapper around #InsertR
		Insert(point int, svalue string)

		// Returns the string of the specified Region.
		// Typically just a wrapper around #SubstrR
		Substr(r Region) string

		ChangeCount() int
		// Returns the line region at the given offset
		Line(offset int) Region
		// Returns a Region starting at the start of a line and ending at the end of a (possibly different) line
		LineR(r Region) Region
		// Returns the lines intersecting the region
		Lines(r Region) []Region
		// Like #Line, but includes the line endings
		FullLine(offset int) Region
		// Like #LineR, but includes the line endings
		FullLineR(r Region) Region
		// Returns the word region at the given text offset
		Word(offset int) Region
		// Returns the Region covering the start of the word in r.Begin()
		// to the end of the word in r.End()
		WordR(r Region) Region
	}

	// The BufferChangedCallback is called everytime a buffer is
	// changed.
	BufferChangedCallback func(buf Buffer, position, delta int)

	buffer struct {
		HasId
		HasSettings
		SerializedBuffer
		changecount int
		name        string
		filename    string
		callbacks   []BufferChangedCallback

		sync.Mutex            // Change lock
		lock       sync.Mutex // All lock
	}
)

// Returns a new empty Buffer
func NewBuffer() Buffer {
	b := buffer{}
	b.SerializedBuffer.init(&rebalancingNode{})
	r := &b
	runtime.SetFinalizer(r, func(b *buffer) { b.Close() })

	return r
}

func (b *buffer) AddCallback(cb BufferChangedCallback) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.callbacks = append(b.callbacks, cb)
}

func (b *buffer) SetName(n string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.name = n
}

func (b *buffer) Name() string {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.name
}

func (b *buffer) FileName() string {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.filename
}

func (b *buffer) SetFileName(n string) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.filename = n
}

func (buf *buffer) notify(position, delta int) {
	for i := range buf.callbacks {
		buf.callbacks[i](buf, position, delta)
	}
}

func (buf *buffer) Insert(point int, svalue string) {
	if len(svalue) == 0 {
		return
	}
	value := []rune(svalue)
	buf.SerializedBuffer.InsertR(point, value)
	buf.Lock()
	defer buf.Unlock()
	buf.lock.Lock()
	defer buf.lock.Unlock()
	buf.changecount++
	buf.notify(point, len(value))
}

func (buf *buffer) Erase(point, length int) {
	if length == 0 {
		return
	}
	buf.Lock()
	defer buf.Unlock()
	buf.lock.Lock()
	defer buf.lock.Unlock()
	buf.changecount++
	buf.SerializedBuffer.Erase(point, length)
	buf.notify(point+length, -length)
}

func (b *buffer) Substr(r Region) string {
	return string(b.SubstrR(r))
}

func (b *buffer) ChangeCount() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.changecount
}

func (b *buffer) Line(offset int) Region {
	if offset < 0 {
		return Region{0, 0}
	} else if s := b.Size(); offset >= s {
		return Region{s, s}
	}
	soffset := offset
sloop:
	o := Clamp(0, soffset, soffset-32)
	sub := b.SubstrR(Region{o, soffset})
	s := soffset
	for s > o && sub[s-o-1] != '\n' {
		s--
	}
	if s == o && o > 0 && sub[0] != '\n' {
		soffset = o
		goto sloop
	}

	l := b.Size()
	eoffset := offset
eloop:
	o = Clamp(eoffset, l, eoffset+32)
	sub = b.SubstrR(Region{eoffset, o})
	e := eoffset
	for e < o && sub[e-eoffset] != '\n' {
		e++
	}
	if e == o && o < l && sub[o-eoffset-1] != '\n' {
		eoffset = o
		goto eloop
	}
	return Region{s, e}
}

func (b *buffer) Lines(r Region) (lines []Region) {
	r = b.LineR(r)
	buf := b.SubstrR(r)
	last := r.Begin()
	for i, ru := range buf {
		if ru == '\n' {
			lines = append(lines, Region{last, r.Begin() + i})
			last = r.Begin() + i + 1
		}
	}
	if last != r.End() {
		lines = append(lines, Region{last, r.End()})
	}
	return lines
}

func (b *buffer) LineR(r Region) Region {
	s := b.Line(r.Begin())
	e := b.Line(r.End())
	return Region{s.Begin(), e.End()}
}

func (b *buffer) FullLine(offset int) Region {
	r := b.Line(offset)
	s := b.Size()
	for r.B < s {
		if i := b.Index(r.B); i == '\r' || i == '\n' {
			break
		}
		r.B++
	}
	if r.B != b.Size() {
		r.B++
	}
	return r
}

func (b *buffer) FullLineR(r Region) Region {
	s := b.FullLine(r.Begin())
	e := b.FullLine(r.End())
	return Region{s.Begin(), e.End()}
}

func (b *buffer) Word(offset int) Region {
	if offset < 0 {
		offset = 0
	}
	lr := b.FullLine(offset)
	col := offset - lr.Begin()

	line := b.SubstrR(lr)
	if len(line) == 0 {
		return Region{offset, offset}
	}

	seps := "./\\()\"'-:,.;<>~!@#$%^&*|+=[]{}`~?"
	if v, ok := b.Settings().Get("word_separators", seps).(string); ok {
		seps = v
	}
	spacing := " \n\t\r"
	eseps := seps + spacing

	if col >= len(line) {
		col = len(line) - 1
	}
	last := true
	li := 0
	ls := false
	lc := 0
	for i, r := range line {
		cur := strings.ContainsRune(eseps, r)
		cs := r == ' '
		if !cs {
			lc = i
		}
		if last == cur && ls == cs {
			continue
		}
		ls = cs
		r := Region{li, i}
		if r.Contains(col) && i != 0 {
			r.A, r.B = r.A+lr.Begin(), r.B+lr.Begin()
			if !(r.B == offset && last) {
				return r
			}
		}
		li = i
		last = cur
	}
	r := Region{lr.Begin() + li, lr.End()}
	lc += lr.Begin()
	if lc != offset && !strings.ContainsRune(spacing, b.Index(r.A)) {
		r.B = lc
	}
	if r.A == offset && r.B == r.A+1 {
		r.B--
	}
	return r
}

// Returns a region that starts at the first character in a word
// and ends with the last character in a (possibly different) word
func (b *buffer) WordR(r Region) Region {
	s := b.Word(r.Begin())
	e := b.Word(r.End())
	return Region{s.Begin(), e.End()}
}
