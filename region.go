package text

import (
	"fmt"
)

type (
	// Defines a Region from point A to B.
	// A can be less than B, in which case
	// the selection is inverted.
	Region struct {
		A, B int
	}
)

func (r Region) String() string {
	return fmt.Sprintf("(%d, %d)", r.A, r.B)
}

// Returns the starting point of the region,
// that would be whichever of A and B
// is the minimal value.
func (r Region) Begin() int {
	return Min(r.A, r.B)
}

// Returns the ending point of the region,
// that would be whichever of A and B
// is the maximum value.
func (r Region) End() int {
	return Max(r.A, r.B)
}

// Returns whether the region contains the given
// point or not.
func (r Region) Contains(point int) bool {
	return point >= r.Begin() && point <= r.End()
}

// Returns whether the region fully covers the argument region
func (r Region) Covers(r2 Region) bool {
	return r.Contains(r2.Begin()) && r2.End() <= r.End()
}

// Returns whether the region is empty or not
func (r Region) Empty() bool {
	return r.A == r.B
}

// Returns the size of the region
func (r Region) Size() int {
	return r.End() - r.Begin()
}

// Returns a region covering both regions
func (r Region) Cover(other Region) Region {
	return Region{Min(r.Begin(), other.Begin()), Max(r.End(), other.End())}
}

// Clips this Region against the Region provided in the argument.
func (r Region) Clip(other Region) Region {
	return Region{Clamp(other.Begin(), other.End(), r.A), Clamp(other.Begin(), other.End(), r.B)}
}

// Returns whether the two regions intersects or not
func (r Region) Intersects(other Region) bool {
	return r == other || r.Intersection(other).Size() > 0
}

// Returns the Region that is the intersection of the two
// regions given
func (r Region) Intersection(other Region) (ret Region) {
	if r.Contains(other.Begin()) || other.Contains(r.Begin()) {
		r2 := Region{Max(r.Begin(), other.Begin()), Min(r.End(), other.End())}
		if r2.Size() != 0 {
			ret = r2
		}
	}

	return ret
}
