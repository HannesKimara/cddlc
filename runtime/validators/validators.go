package validators

import (
	"reflect"
	"regexp"

	"golang.org/x/exp/constraints"
)

// MunericOrdered is a constraint that allows either integer or floating point type.
type NumericOrdered interface {
	constraints.Integer | constraints.Float
} // This is used instead of constraints.Ordered since cddl ranges currently only support
// integers or floats

// InExclusive range return true if the val is within the range (lower, upper)
// The lower and upper values are not inlcuded
func InExclusiveRange[T NumericOrdered](val, lower, upper T) bool {
	return val > lower && val < upper
}

// InInclusiveRange returns true if the val is within the range[lower, upper]
// The value may be equal to the limits
func InInclusiveRange[T NumericOrdered](val, lower, upper T) bool {
	return val >= lower && val <= upper
}

// IsLessThan returns true if the val is less than the upper limit
func IsLessThan[T NumericOrdered](val, limit T) bool {
	return val < limit
}

// IsGreaterThan returns true if the val is greater than the upper bound
func IsGreaterThan[T NumericOrdered](val, limit T) bool {
	return val > limit
}

// RegexpMatches validates a string against a compiled regexp instance.
func RegexpMatches(regex *regexp.Regexp, input string) bool {
	return regex.Match([]byte(input))
}

// CheckSize returns whether the size of the value in bytes is less than the upper bound
func CheckSize(val interface{}, size uint) bool {
	return reflect.TypeOf(val).Size() < uintptr(size)
}

func MustBool(out bool) {
	if !out {
		panic("received false")
	}
}

func MustNotErr(err error) {
	if err != nil {
		panic(err)
	}
}
