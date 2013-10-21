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
	if !reflect.DeepEqual(r.regions, []Region{{5, 10}, {10, 23}}) {
		t.Errorf("Not as expected: %v", r)
	}

	r.Add(Region{2, 6})
	if !reflect.DeepEqual(r.regions, []Region{{2, 10}, {10, 23}}) {
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
