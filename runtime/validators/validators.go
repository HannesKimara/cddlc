package validators

import (
	"golang.org/x/exp/constraints"
)

// RangeOrdered is a constraint that permits either integer of floating point type.
//
// This is used instead of constraints.Ordered since cddl ranges currently only support
// integers or floats
type RangeOrdered interface {
	constraints.Integer | constraints.Float
}

func InExclusiveRange[T RangeOrdered](val, lower, upper T) bool {
	return val > lower && val < upper
}

func InInclusiveRange[T RangeOrdered](val, lower, upper T) bool {
	return val >= lower && val <= upper
}
