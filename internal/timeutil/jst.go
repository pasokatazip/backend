package timeutil

import "time"

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

// NowJST returns the current time in Japan Standard Time.
func NowJST() time.Time {
	return time.Now().In(jst)
}

// LocationJST returns the Japan Standard Time location.
func LocationJST() *time.Location {
	return jst
}
