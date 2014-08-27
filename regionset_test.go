// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"reflect"
	"testing"
)

// Verified against ST3
func TestRegionSetAdjust(t *testing.T) {
	var r RegionSet

	r.AddAll([]Region{
		{10, 20},
		{25, 35},
	})

	r.Adjust(2, 5)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {30, 40}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Adjust(30, 1)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 41}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Adjust(41, 1)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 42}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Adjust(43, 1)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 42}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Adjust(44, -5)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 39}}) {
		t.Errorf("Not as expected: %v", r)
	}
	r.Adjust(44, -5)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 39}}) {
		t.Errorf("Not as expected: %v", r)
	}
	r.Adjust(43, -5)
	if !reflect.DeepEqual(r.regions, []Region{{15, 25}, {31, 38}}) {
		t.Errorf("Not as expected: %v", r)
	}
}

// Verified against ST3
func TestRegionSetflush(t *testing.T) {
	var r RegionSet
	r.Add(Region{10, 20})
	r.Add(Region{15, 23})
	if !reflect.DeepEqual(r.regions, []Region{{10, 23}}) {
		t.Errorf("Not as expected: %v", r)
	}
	r.Add(Region{5, 10})
	if !reflect.DeepEqual(r.regions, []Region{{10, 23}, {5, 10}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Add(Region{2, 6})
	if !reflect.DeepEqual(r.regions, []Region{{10, 23}, {2, 10}}) {
		t.Errorf("Not as expected: %v", r)
	}
	r.Clear()
	r.Add(Region{10, 10})
	r.Add(Region{10, 11})
	if !reflect.DeepEqual(r.regions, []Region{{10, 11}}) {
		t.Errorf("Not as expected: %v", r)
	}
}

// Verified against ST3
func TestRegionSetAdjust2(t *testing.T) {
	var r RegionSet

	r.AddAll([]Region{
		{10, 20},
		{25, 35},
	})

	r.Adjust(43, -25)
	if !reflect.DeepEqual(r.regions, []Region{{10, 18}, {18, 18}}) {
		t.Errorf("Not as expected: %v", r)
	}
}

func TestRegionSetCut(t *testing.T) {
	tests := []struct {
		A, B Region
		Out  RegionSet
	}{
		{Region{10, 20}, Region{0, 5}, RegionSet{regions: []Region{{10, 20}}}},
		{Region{10, 20}, Region{12, 15}, RegionSet{regions: []Region{{10, 12}, {15, 20}}}},
		{Region{10, 20}, Region{5, 15}, RegionSet{regions: []Region{{15, 20}}}},
		{Region{10, 20}, Region{15, 20}, RegionSet{regions: []Region{{10, 15}}}},
	}
	for i, test := range tests {
		var rs RegionSet
		rs.Add(test.A)
		t.Log(rs)
		if res := rs.Cut(test.B); !reflect.DeepEqual(res, test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, res)
		}
	}
}

func TestRegionSetAdd(t *testing.T) {
	tests := []struct {
		A   []Region
		B   Region
		Out []Region
	}{
		{[]Region{{10, 20}}, Region{0, 5}, []Region{{10, 20}, {0, 5}}},
		{[]Region{{10, 20}}, Region{12, 15}, []Region{{10, 20}}},
		{[]Region{{10, 20}}, Region{5, 15}, []Region{{5, 20}}},
		{[]Region{{10, 20}}, Region{15, 25}, []Region{{10, 25}}},
		{[]Region{{10, 20}}, Region{20, 25}, []Region{{10, 20}, {20, 25}}},
		{[]Region{{10, 15}, {20, 25}}, Region{12, 23}, []Region{{10, 25}}},
	}
	for i, test := range tests {
		var v RegionSet
		v.AddAll(test.A)
		v.Add(test.B)
		if !reflect.DeepEqual(v.Regions(), test.Out) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.Out, v.Regions())
		}
	}
}

func TestRegionSetAddAll(t *testing.T) {
	tests := []struct {
		in  []Region
		exp []Region
	}{
		{
			[]Region{{5, 15}, {0, 20}, {100, 90}, {10, 25}, {45, 30}},
			[]Region{{0, 25}, {100, 90}, {45, 30}},
		},
		{
			[]Region{{100, 50}, {20, 5}, {0, 10}, {30, 40}, {15, 25}},
			[]Region{{100, 50}, {25, 0}, {30, 40}},
		},
	}
	for i, test := range tests {
		var v RegionSet
		v.AddAll(test.in)
		if !reflect.DeepEqual(v.Regions(), test.exp) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.exp, v.Regions())
		}
	}
}

func TestRegionSubtract(t *testing.T) {
	tests := []struct {
		A      []Region
		B      Region
		expect []Region
	}{
		{
			[]Region{{1, 4}, {6, 10}, {15, 25}},
			Region{6, 10},
			[]Region{{1, 4}, {15, 25}},
		},
		{
			[]Region{{6, 10}, {15, 25}},
			Region{7, 9},
			[]Region{{6, 7}, {9, 10}, {15, 25}},
		},
	}
	for i, test := range tests {
		var v RegionSet
		v.AddAll(test.A)
		v.Substract(test.B)
		if !reflect.DeepEqual(v.Regions(), test.expect) {
			t.Errorf("Test %d; Expected %v, got: %v", i, test.expect, v.Regions())
		}
	}
}
