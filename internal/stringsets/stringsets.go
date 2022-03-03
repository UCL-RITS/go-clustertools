// A Golang implementation of the Python3 set API, for only sets of strings.

package stringsets

import (
	"errors"
)

type StringSet struct {
	data map[string]bool
}

// Some of list of functions cribbed from Python3's set class

// Create a new StringSet pointer properly.
func New() *StringSet {
	s := StringSet{}
	s.data = make(map[string]bool)
	return &s
}

func NewFromSlice(s []string) *StringSet {
	set := New()
	for _, v := range s {
		set.Add(v)
	}
	return set
}

// Returns true if the element is in the set.
func (set *StringSet) Has(s string) bool {
	if set.data[s] == true {
		return true
	}
	return false
}

// Returns true if two sets have the same contents.
func (set *StringSet) Equals(set2 *StringSet) bool {
	if len(set.data) != len(set2.data) {
		return false
	}
	for k, _ := range set.data {
		if !set2.Has(k) {
			return false
		}
	}
	return true
}

// Returns the set contents as a slice of strings.
func (set *StringSet) AsSlice() []string {
	slice := []string{}
	for k, _ := range set.data {
		slice = append(slice, k)
	}
	return slice
}

// Returns true if each element of the slice is in the set, and no others.
func (set *StringSet) EqualsSlice(s []string) bool {
	s2 := NewFromSlice(s)
	return set.Equals(s2)
}

// Returns number of elements of the set.
func (set *StringSet) Len() int {
	return len(set.data)
}

// Add an element to a set.
//
// This has no effect if the element is already present.
func (set *StringSet) Add(s string) {
	set.data[s] = true
}

// Adds each element of a slice to a set.
func (set *StringSet) AddSlice(s []string) {
	for _, v := range s {
		set.Add(v)
	}
}

// Remove all elements from this set.
func (set *StringSet) Clear() {
	for k, _ := range set.data {
		delete(set.data, k)
	}
}

// Return a shallow copy of a set.
func (set *StringSet) Copy() *StringSet {
	s2 := New()
	for k, _ := range set.data {
		s2.Add(k)
	}
	return s2
}

// Return the difference of two or more sets as a new set.
//
// (i.e. all elements that are in this set but not the others.)
func (set *StringSet) Difference(set2 *StringSet) *StringSet {
	s3 := set.Copy()
	s3.DifferenceUpdate(set2)
	return s3
}

// Remove all elements of another set from this set.
func (set *StringSet) DifferenceUpdate(set2 *StringSet) {
	for k, _ := range set2.data {
		set.Discard(k)
	}
}

// Remove an element from a set if it is a member.
//
// If the element is not a member, do nothing.
func (set *StringSet) Discard(s string) {
	// "If [the first arg] is nil or there is no such element, delete is a no-op."
	delete(set.data, s)
}

// Remove each element of a slice from a set.
func (set *StringSet) DiscardSlice(s []string) {
	for _, v := range s {
		set.Discard(v)
	}
}

// Return the intersection of two sets as a new set.
//
// (i.e. all elements that are in both sets.)
func (set *StringSet) Intersection(set2 *StringSet) *StringSet {
	s3 := set.Copy()
	s3.IntersectionUpdate(set2)
	return s3
}

// Update a set with the intersection of itself and another.
func (set *StringSet) IntersectionUpdate(set2 *StringSet) {
	for k, _ := range set.data {
		if !set2.Has(k) {
			set2.Discard(k)
		}
	}
}

// Return true if set has a null intersection with set 2 (i.e. they share no members).
func (set *StringSet) IsDisjoint(set2 *StringSet) bool {
	s3 := set.Intersection(set2)
	if len(s3.data) == 0 {
		return true
	}
	return false
}

// Return true if set2 contains set.
func (set *StringSet) IsSubset(set2 *StringSet) bool {
	for k, _ := range set.data {
		if !set2.Has(k) {
			return false
		}
	}
	return true
}

// Return true if set contains set2.
func (set *StringSet) IsSuperset(set2 *StringSet) bool {
	return set2.IsSubset(set)
}

// Remove and return an arbitrary set element.
//
// Returns empty string and error if the set is empty.
func (set *StringSet) Pop() (string, error) {
	for k, _ := range set.data {
		set.Discard(k)
		return k, nil
	}
	return "", errors.New("no elements remaining")
}

// Remove an element from a set; it must be a member.
//
// If the element is not a member, return error.
func (set *StringSet) Remove(k string) error {
	if set.Has(k) {
		set.Discard(k)
		return nil
	}
	return errors.New("key is not in set")
}

// Return the symmetric difference of two sets as a new set.
//
// (i.e. all elements that are in exactly one of the sets.)
func (set *StringSet) SymmetricDifference(set2 *StringSet) *StringSet {
	s3 := set.Copy()
	s3.SymmetricDifferenceUpdate(set2)
	return s3
}

// Update a set with the symmetric difference of itself and another.
func (set *StringSet) SymmetricDifferenceUpdate(set2 *StringSet) {
	s3 := set.Copy()
	s3.IntersectionUpdate(set2)
	set.UnionUpdate(set2)
	set.DifferenceUpdate(s3)
}

// Return the union of sets as a new set.
//
// (i.e. all elements that are in either set.)
func (set *StringSet) Union(set2 *StringSet) *StringSet {
	s3 := set.Copy()
	s3.UnionUpdate(set2)
	return s3
}

// Update a set with the union of itself and another set.
func (set *StringSet) UnionUpdate(set2 *StringSet) {
	for k, _ := range set2.data {
		set.Add(k)
	}
}
