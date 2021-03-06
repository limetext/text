// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	//	"math/rand"
	"reflect"
	"testing"
)

var complexnode_test = &node{
	22, 0,
	&node{
		9, 0,
		&node{
			9, 0,
			&node{
				6, 0,
				&node{6, 0, nil, nil, []rune("Hello ")},
				&node{3, 0, nil, nil, []rune("my ")},
				nil,
			},
			nil,
			nil,
		},
		&node{
			7, 0,
			&node{
				6, 0,
				&node{
					2, 0,
					&node{2, 0, nil, nil, []rune("na")},
					&node{4, 0, nil, nil, []rune("me i")},
					nil,
				},
				&node{1, 0, nil, nil, []rune("s")},
				nil,
			},
			&node{6, 0, nil, nil, []rune(" Simon")},
			nil,
		},
		nil,
	},
	nil,
	nil,
}

type Test struct {
	in  *node
	exp string
}

var (
	tests = []Test{
		{&node{6, 0, &node{6, 0, nil, nil, []rune("Hello ")}, &node{5, 0, nil, nil, []rune("world")}, nil}, "Hello world"},
		{&node{6, 0, &node{6, 0, nil, nil, []rune("Hello ")}, &node{3, 0, &node{3, 0, nil, nil, []rune("wor")}, &node{2, 0, nil, nil, []rune("ld")}, nil}, nil}, "Hello world"},
		{&node{6, 0, &node{6, 0, nil, nil, []rune("Hello ")}, &node{5, 0, nil, nil, []rune("world")}, nil}, "Hello world"},
		{complexnode_test, "Hello my name is Simon"},
	}
	merges = []int{4, 8, 32, 128, 1024, merge}
)

func init() {
	const (
		size  = 1024
		split = 8
	)
	in := make([]rune, size)
	fill(in)
	tests = append(tests, Test{newNodeEx(in, split), string(in)})
}

func TestNode(t *testing.T) {
	for i, test := range tests {
		if sub := test.in.Substr(Region{0, len(test.exp)}); sub != test.exp {
			t.Fatalf("%d %s != %s", i, sub, test.exp)
		} else if l := test.in.Size(); l != len(sub) {
			t.Fatalf("%d %d != %d", i, l, len(sub))
		}
	}
}

func TestNodeSimplify(t *testing.T) {
	r := &node{5, 0, nil, nil, []rune("world")}
	l := &node{0, 0, nil, nil, nil}
	n := node{0, 0, l, r, nil}
	n.simplify()
	if !reflect.DeepEqual(&n, r) {
		t.Error(n.dump(""))
	}
	n = node{5, 0, r, l, nil}
	n.simplify()
	if !reflect.DeepEqual(&n, r) {
		t.Error(n.dump(""))
	}
}

func TestNodeSplit(t *testing.T) {
	for _, m := range merges {
		merge = m
		for i, test := range tests {
			for j := range test.exp {
				nn := test.in.clone()
				r := nn.split(j)
				if sub := nn.Substr(Region{0, j}) + r.Substr(Region{0, len(test.exp) - j}); sub != test.exp {
					t.Fatalf("%d, %d, split %s != %s:\n%s\n%s", i, j, sub, test.exp, nn.dump("\t"), r.dump("\t"))
				} else if l := nn.Size(); l != j {
					t.Fatalf("%d, %d, split length1 %d != %d:\n%s\n%s", i, j, l, j, nn.dump("\t"), r.dump("\t"))
				} else if l := r.Size(); l != len(test.exp)-j {
					t.Fatalf("%d, %d, split length2 %d != %d:\n%s\n", i, j, l, len(test.exp)-j, r.dump("\t"))
				}
			}
		}
	}
}

func TestNodeConcat(t *testing.T) {
	for _, m := range merges {
		merge = m
		for i, test := range tests {
			for j := range test.exp {
				nn := test.in.clone()
				r := nn.split(j)
				nn.concat(r)
				if sub := nn.Substr(Region{0, len(test.exp)}); sub != test.exp {
					t.Fatalf("%d, %d, split/concat %s != %s:\n%s", i, j, sub, test.exp, nn.dump("\t"))
				} else if l := nn.Size(); l != len(test.exp) {
					t.Fatalf("%d, %d, %d != %d", i, j, l, len(test.exp))
				}
			}
		}
	}
}

func TestNodeInsertR(t *testing.T) {
	if testing.Short() {
		t.Skip("Short")
	}

	const (
		size  = 256
		isize = 32
	)
	od := make([]rune, size)
	fill(od)

	in := make([]rune, size)
	fill(in)

	for _, m := range merges {
		merge = m
		for i := range od {
			n := newNode(od)
			b := naiveBuffer{}
			b.InsertR(0, od)
			n.InsertR(i, in)
			b.InsertR(i, in)
			r := Region{0, b.Size()}

			if b.Size() != n.Size() {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %d != %d\n%s", m, i, b.Size(), n.Size(), na)
			} else if e, a := string(b.SubstrR(r)), string(n.SubstrR(r)); e != a {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %s != %s\n%s", m, i, e, a, na)
			}
		}
		for i := range od {
			n := newNode(od)
			for _, j := range in {
				l := n.Size()
				n.InsertR(i, []rune{j})
				if n.Size() != l+1 {
					t.Log(string(j))
					na := n.dump("\t")
					t.Fatalf("%d, %d, %d: %d != %d\n%s", m, i, j, n.Size(), l+1, na)
				}
			}
		}
	}

	offs := []int{-100, len(od) - 1, len(od) + 1, len(od) + 100}

	for _, m := range merges {
		for i, o := range offs {
			n := newNode(od)
			b := naiveBuffer{}
			b.InsertR(0, od)
			n.InsertR(o, in)
			b.InsertR(o, in)
			r := Region{0, b.Size()}

			if b.Size() != n.Size() {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %d != %d\n%s", m, i, b.Size(), n.Size(), na)
			} else if e, a := string(b.SubstrR(r)), string(n.SubstrR(r)); e != a {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %s != %s\n%s", m, i, e, a, na)
			}
		}
	}
}

func TestNodeAppend(t *testing.T) {
	if testing.Short() {
		t.Skip("Short")
	}

	const (
		size   = 1024 * 4
		isize  = 256
		target = 1024 * 1024
	)
	od := make([]rune, size)
	fill(od)

	in := make([]rune, size)
	fill(in)

	for _, m := range merges {
		merge = m
		n := newNode(od)
		for n.Size() < target {
			n.InsertR(n.Size(), in)
		}
	}
}

func TestNodeRowCol(t *testing.T) {
	if testing.Short() {
		t.Skip("Short")
	}
	const (
		size  = 256
		isize = 256
	)
	od := make([]rune, size)
	fill(od)

	in := make([]rune, size)
	fill(in)

	for _, m := range merges {
		merge = m
		n := newNode(od)
		b := naiveBuffer{}
		b.InsertR(0, od)
		for i := 0; i < 30; i++ {
			n.InsertR(n.Size(), in)
			b.InsertR(b.Size(), in)
			r := Region{0, b.Size()}
			if b.Size() != n.Size() {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %d != %d\n%s", m, i, b.Size(), n.Size(), na)
			} else if e, a := string(b.SubstrR(r)), string(n.SubstrR(r)); e != a {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %s != %s\n%s", m, i, e, a, na)
			} else {
				for j := -5; j < n.Size()+10; j++ {
					r1, c1 := n.RowCol(j)
					r2, c2 := b.RowCol(j)
					if r1 != r2 || c1 != c2 {
						t.Fatalf("%d, %d: %d != %d || %d != %d", m, i, r1, r2, c1, c2)
					} else if tp1, tp2 := b.TextPoint(r2, c2), n.TextPoint(r1, c1); tp1 != tp2 {
						t.Fatalf("%d, %d: %d != %d", m, i, tp1, tp2)
					}
				}
			}
		}
	}
}

func TestNodeSubstr(t *testing.T) {
	if testing.Short() {
		t.Skip("Short")
	}
	const (
		size  = 256
		isize = 13
	)
	od := make([]rune, size)
	fill(od)

	in := make([]rune, size)
	fill(in)

	for _, m := range merges {
		merge = m
		n := newNode(od)
		b := naiveBuffer{}
		b.InsertR(0, od)
		for i := 0; i < 10; i++ {
			n.InsertR(n.Size(), in)
			b.InsertR(b.Size(), in)
			if b.Size() != n.Size() {
				na := n.dump("\t")
				t.Fatalf("%d, %d: %d != %d\n%s", m, i, b.Size(), n.Size(), na)
			}
			regions := []Region{
				{0, b.Size()},
				{b.Size() / 2, b.Size()},
				{b.Size() / 4, b.Size() / 2},
				{b.Size() / 8, b.Size() / 4},
				{b.Size() / 16, b.Size() / 8},
				{b.Size() - b.Size()/32, b.Size()},
				{b.Size() - b.Size()/16, b.Size()},
				{b.Size() - b.Size()/8, b.Size()},
				{b.Size() - b.Size()/4, b.Size()},
				{b.Size() - b.Size()/2, b.Size()},
				{b.Size()/2 - b.Size()/32, b.Size() / 2},
				{b.Size()/2 - b.Size()/16, b.Size() / 2},
				{b.Size()/2 - b.Size()/8, b.Size() / 2},
				{b.Size()/2 - b.Size()/4, b.Size() / 2},
				{b.Size()/4 - b.Size()/32, b.Size() / 4},
				{b.Size()/4 - b.Size()/16, b.Size() / 4},
				{b.Size()/4 - b.Size()/8, b.Size() / 4},
				{b.Size()/4 - b.Size()/8, b.Size() + 10},
				{b.Size() - 10, b.Size() + 10},
				{-10, 10},
			}
			for _, r := range regions {
				if e, a := string(b.SubstrR(r)), string(n.SubstrR(r)); e != a {
					na := n.dump("\t")
					t.Fatalf("%d, %d: %s != %s\n%s", m, i, e, a, na)
				}
			}
		}
	}
}

func TestNodeErase(t *testing.T) {
	if testing.Short() {
		t.Skip("Short")
	}

	const (
		size  = 2 * 1024
		dsize = 1
	)
	od := make([]rune, size)
	fill(od)

	for _, m := range merges {
		merge = m
		for i := range od {
			n := newNode(od)
			b := naiveBuffer{}
			b.InsertR(0, od)
			n.Erase(i, dsize)
			b.Erase(i, dsize)
			r := Region{0, b.Size()}

			if b.Size() != n.Size() {
				t.Fatalf("%d, %d: %d != %d\n", m, i, b.Size(), n.Size())
			} else if e, a := string(b.SubstrR(r)), string(n.SubstrR(r)); e != a {
				r = Region{0, 20}
				e = string(b.SubstrR(r))
				a = string(n.SubstrR(r))
				t.Fatalf("%d, %d: %s != %s (%v)", m, i, e, a, e != a)
			}
		}
	}
}

// func BenchmarkNodeSplit(b *testing.B) {
// 	b.StopTimer()
// 	data := make([]rune, 1024*256)
// 	fill(data)
// 	buf := newNodeEx(data, 4096)
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		b2 := buf.clone()
// 		b2.split(512)
// 	}
// }

// func BenchmarkNodeInsertRand(b *testing.B) {
// 	r := rand.Perm(b.N)
// 	b.StopTimer()
// 	data := []rune(testinsert())
// 	buf := newNode(testbuffer().Runes())
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		l := buf.Size()
// 		pos := r[i] % l
// 		buf.InsertR(pos, data)
// 	}
// 	buf.Substr(Region{0, buf.Size()})
// }

// func BenchmarkNodeInsertBegin(b *testing.B) {
// 	b.StopTimer()
// 	sdata := testinsert()
// 	in := testbuffer().Runes()
// 	buf := newNode(in)
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.InsertR(0, sdata)
// 	}
// 	buf.Substr(Region{0, buf.Size()})
// 	if a, e := buf.Size(), b.N*len(sdata)+len(in); a != e {
// 		b.Error(a, e)
// 	}
// }

// func BenchmarkNodeInsertMid(b *testing.B) {
// 	b.StopTimer()
// 	sdata := testinsert()
// 	in := testbuffer().Runes()
// 	buf := newNode(in)
// 	b.StartTimer()

// 	for i := 0; i < b.N; i++ {
// 		buf.InsertR(buf.Size()/2, sdata)
// 	}
// 	buf.Substr(Region{0, buf.Size()})
// 	if a, e := buf.Size(), b.N*len(sdata)+len(in); a != e {
// 		b.Error(a, e)
// 	}
// }

// func BenchmarkNodeInsertEnd(b *testing.B) {
// 	b.StopTimer()
// 	sdata := testinsert()
// 	in := testbuffer().Runes()
// 	buf := newNode(in)
// 	b.StartTimer()
// 	for i := 0; i < b.N; i++ {
// 		buf.InsertR(buf.Size(), sdata)
// 	}
// 	buf.Substr(Region{0, buf.Size()})
// 	if a, e := buf.Size(), b.N*len(sdata)+len(in); a != e {
// 		b.Error(a, e)
// 	}
// }

// func BenchmarkNodeRune(b *testing.B) {
// 	buf := newNode(testbuffer().Runes())
// 	r := rand.Perm(b.N)
// 	l := buf.Size()
// 	for i := 0; i < b.N; i++ {
// 		buf.Rune(r[i] % l)
// 	}
// }
