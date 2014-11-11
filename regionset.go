// Copyright 2013 Fredrik Ehnbom
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package text

import (
	"sync"
)

// The RegionSet manages multiple regions,
// merging any regions overlapping.
//
// Note that regions that are right next to each other
// are not merged into a single region. This is because
// otherwise it would not be possible to have multiple
// cursors right next to each other.
type RegionSet struct {
	regions []Region
	lock    sync.Mutex
}

// Adjusts all the regions in the set
func (r *RegionSet) Adjust(position, delta int) {
	r.lock.Lock()
	defer r.lock.Unlock()
	for i := range r.regions {
		r.regions[i].Adjust(position, delta)
	}
	r.flush()
}

// Returns a list of the indices between start and end of the regions that overlaps
// with the given reference region.
func (r *RegionSet) overlaps(reference Region, start, end int) (ret []int) {
	for i := start; i < end; i++ {
		if reference == r.regions[i] || reference.Intersects(r.regions[i]) || reference.Covers(r.regions[i]) {
			ret = append(ret, i)
		}
	}
	return
}

// Merge all regions in the given "merge"-list with the region at index "reference"
func (r *RegionSet) merge(reference int, merge []int) {
	for _, j := range merge {
		// merge "j" into "reference"
		r.regions[reference] = r.regions[reference].Cover(r.regions[j])
	}
	l := len(merge) - 1
	// keep track of how many indices we have removed thus far
	adj := 0
	for i, j1 := range merge {
		j2 := len(r.regions) - adj
		if i < l {
			j2 = merge[i+1] - 1
		}
		// remove "j" from the region list by shifting all trailing regions up one step
		if j2 > 0 && j1+1 <= j2 {
			copy(r.regions[j1-adj:], r.regions[j1+1:j2])
		}
		adj++
	}
	r.regions = r.regions[:len(r.regions)-len(merge)]
}

// TODO(q): There should be a on modified callback on the RegionSet
func (r *RegionSet) flush() {
	for i := 1; i < len(r.regions); i++ {
		ov := r.overlaps(r.regions[i], 0, i)
		if len(ov) == 0 {
			continue
		}
		r.merge(ov[0], append(ov[1:], i))
	}
}

// Removes the given region from the set
func (r *RegionSet) Substract(r2 Region) {
	r3 := r.Cut(r2)
	r.lock.Lock()
	defer r.lock.Unlock()
	r.regions = r3.regions
}

// Adds the given region to the set
func (r *RegionSet) Add(r2 Region) {
	r.lock.Lock()
	defer r.lock.Unlock()

	ov := r.overlaps(r2, 0, len(r.regions))
	r.regions = append(r.regions, r2)
	if len(ov) == 0 {
		return
	}
	ref := ov[0]
	ov = append(ov[1:], len(r.regions)-1)
	r.merge(ref, ov)
}

// Clears the set
func (r *RegionSet) Clear() {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.regions = r.regions[0:0]
	r.flush()
}

// Gets the region at index i
func (r *RegionSet) Get(i int) Region {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.regions[i]
}

// Returns the number of regions contained in the set
func (r *RegionSet) Len() int {
	return len(r.regions)
}

// Adds all regions in the array to the set
func (r *RegionSet) AddAll(rs []Region) {
	r.lock.Lock()
	defer r.lock.Unlock()
	// Merge regions in rs that overlap
	rr := RegionSet{regions: rs}
	rr.flush()
	rs = rr.Regions()

	// r.regions is already by itself maintained
	// as a non-overlapping RegionSet
	start := len(r.regions)
	r.regions = append(r.regions, rs...)

	// In other words, we just need to check overlap between rs
	// and the previous r.region-set
	for _, r2 := range rs {
		ov := r.overlaps(r2, 0, start)
		if len(ov) == 0 {
			continue
		}
		ref := ov[0]
		ov = append(ov[1:], len(r.regions)-1)
		r.merge(ref, ov)
		start -= len(ov)
	}
}

// Returns whether the specified region is part of the set
func (r *RegionSet) Contains(r2 Region) bool {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i := range r.regions {
		if r.regions[i] == r2 || (r.regions[i].Contains(r2.Begin()) && r.regions[i].Contains(r2.End())) {
			return true
		}
	}
	return false
}

// Returns a copy of the regions in the set
func (r *RegionSet) Regions() (ret []Region) {
	r.lock.Lock()
	defer r.lock.Unlock()
	ret = make([]Region, len(r.regions))
	copy(ret, r.regions)
	return
}

// Returns whether the set contains at least one
// region that isn't empty.
func (r *RegionSet) HasNonEmpty() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, r := range r.regions {
		if !r.Empty() {
			return true
		}
	}
	return false
}

// Opposite of #HasNonEmpty
func (r *RegionSet) HasEmpty() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	for _, r := range r.regions {
		if r.Empty() {
			return true
		}
	}
	return false
}

// Cuts away the provided region from the set, and returns
// the new set
func (r *RegionSet) Cut(r2 Region) (ret RegionSet) {
	r.lock.Lock()
	defer r.lock.Unlock()

	for i := 0; i < len(r.regions); i++ {
		for _, xor := range r.regions[i].Cut(r2) {
			if !xor.Empty() {
				ret.Add(xor)
			}
		}
	}
	return
}
