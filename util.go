package main

import (
	"time"
)

var monthAbbr = map[string]time.Month{
	"jan": 1,
	"feb": 2,
	"mar": 3,
	"apr": 4,
	"may": 5,
	"jun": 6,
	"jul": 7,
	"aug": 8,
	"sep": 9,
	"oct": 10,
	"nov": 11,
	"dec": 12,
}

//StartOfMonth returns a timestamp of the first at 12am of the current month
func StartOfMonth() Timestamp {
	return MonthStart(TimestampNow().Month())
}

func EndOfMonth() Timestamp {
	return MonthEnd(TimestampNow().Month())
}

func MonthStart(m time.Month) Timestamp {
	return Timestamp{time.Date(TimestampNow().Year(), m, 1, 0, 0, 0, 0, time.UTC)}
}

func MonthEnd(m time.Month) Timestamp {
	return Timestamp{time.Date(TimestampNow().Year(), m + 1, 0, 11, 59, 59, 999999999, time.UTC)}
}

func MonthRangeFromString(m string) (*[2]Timestamp, bool) {
	val, ok := monthAbbr[m]
	if !ok {
		return nil, ok
	}
	return monthTimestampRange(val), true
}

func CurrentMonthRange() *[2]Timestamp {
	return monthTimestampRange(TimestampNow().Month())
}

//montTimestampRange returns a slice of two timestamps representing the first nanosecond
//of the month and the last nanosecond of the month
func monthTimestampRange(m time.Month) *[2]Timestamp {
	ts := new([2]Timestamp)
	ts[0] = MonthStart(m)
	ts[1] = MonthEnd(m)
	return ts
}

