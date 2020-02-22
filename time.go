package main

import (
	"encoding/json"
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

//Timestamp is a time.Time with custom json Marshaling/Unmarshaling
type Timestamp time.Time

//TimestampNow returns the current time converted to a
//Timestamp
func TimestampNow() Timestamp {
	return Timestamp(time.Now())
}

//UnMarshallJson implements the json.Unmarshaler interface
//the time is formatted as nanoseconds since the epoch
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var i int64
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	*t = Timestamp(time.Unix(0, i))
	return nil
}

//MarshallJson implements the json.Marshaler interface
//Converts Timestamp into nanoseconds since epoch to be
//marshalled into json
func (t Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time().UnixNano())
}

//String returns the default string representation of the timestamp
//which is the 3 letter day, then 3 letter month then the day of of the month
func (t Timestamp) String() string {
	return time.Time(t).Format("Mon Jan 2")
}

func (t *Timestamp) Time() time.Time {
	return time.Time(*t)
}

func (t *Timestamp) Json() (string, error) {
	return toJsonString(t)
}

//StartOfMonth returns a timestamp of the first at 12am of the current month
func StartOfMonth() time.Time {
	return MonthStart(time.Now().Month())
}

func EndOfMonth() time.Time {
	return MonthEnd(time.Now().Month())
}

func MonthStart(m time.Month) time.Time {
	return time.Date(time.Now().Year(), m, 1, 0, 0, 0, 0, time.UTC)
}

func MonthEnd(m time.Month) time.Time {
	return time.Date(time.Now().Year(), m+1, 0, 11, 59, 59, 999999999, time.UTC)
}

func MonthRangeFromString(m string) (*[2]time.Time, bool) {
	val, ok := monthAbbr[m]
	if !ok {
		return nil, ok
	}
	return monthTimestampRange(val), true
}

func CurrentMonthRange() *[2]time.Time {
	return monthTimestampRange(time.Now().Month())
}

//monthTimestampRange returns a slice of two timestamps representing the first nanosecond
//of the month and the last nanosecond of the month
func monthTimestampRange(m time.Month) *[2]time.Time {
	ts := new([2]time.Time)
	ts[0] = MonthStart(m)
	ts[1] = MonthEnd(m)
	return ts
}
