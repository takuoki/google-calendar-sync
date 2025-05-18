package entity

import "time"

// compareTime is a helper function to compare two *time.Time values.
func compareTime(t1, t2 *time.Time) bool {
	if t1 == nil && t2 == nil {
		return true
	}
	if t1 != nil && t2 != nil {
		return t1.Equal(*t2)
	}

	return false
}

func comparePointer[T comparable](p1, p2 *T) bool {
	if p1 == nil && p2 == nil {
		return true
	}
	if p1 != nil && p2 != nil {
		return *p1 == *p2
	}

	return false
}
