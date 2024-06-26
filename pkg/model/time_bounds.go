package model

import (
	"errors"
	"strconv"
)

// TimeBounds represents a period of time between two timestamps.
type TimeBounds struct {
	FromTime *int64
	ToTime   *int64
}

// NewTimeBounds creates a new TimeBounds using provided fromTime and toTime values.
func NewTimeBounds(fromTime, toTime *int64) (*TimeBounds, error) {
	if fromTime != nil && *fromTime < 0 {
		return nil, errors.New("fromTime should be positive or zero")
	}
	if toTime != nil && *toTime <= 0 {
		return nil, errors.New("toTime should be positive")
	}
	if fromTime != nil && toTime != nil && *fromTime >= *toTime {
		return nil, errors.New("fromTime should be less than toTime")
	}
	return &TimeBounds{FromTime: fromTime, ToTime: toTime}, nil
}

// NewTimeBoundsTo creates a new TimeBounds until the provided toTime.
func NewTimeBoundsTo(toTime int64) (*TimeBounds, error) {
	return NewTimeBounds(nil, &toTime)
}

// NewTimeBoundsFromString creates a TimeBounds from the string representation of
// time boundaries.
func NewTimeBoundsFromString(from, to string) (*TimeBounds, error) {
	fromTime, err := parseTimestamp(from)
	if err != nil {
		return nil, err
	}
	toTime, err := parseTimestamp(to)
	if err != nil {
		return nil, err
	}
	return NewTimeBounds(fromTime, toTime)
}

func parseTimestamp(value string) (*int64, error) {
	if len(value) == 0 {
		return nil, nil
	}
	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// Contains verifies if the given value lies within FromTime (inclusive) and ToTime (exclusive).
func (tb *TimeBounds) Contains(value int64) bool {
	if tb.FromTime != nil && value < *tb.FromTime {
		return false
	}
	if tb.ToTime != nil && value >= *tb.ToTime {
		return false
	}
	return true
}
