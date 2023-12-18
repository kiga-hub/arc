package utils

import "time"

// ParseDateTimeString  "Parse date and time string 2020-05-05 => unix nano
//
//goland:noinspection GoUnusedExportedFunction
func ParseDateTimeString(str string, include bool) (int64, error) {
	if str == "" {
		return 0, nil
	}
	if include {
		str += " 23:59:59"
	} else {
		str += " 00:00:00"
	}
	timeLayout := "2006-01-02 15:04:05"
	loc, err := time.LoadLocation("Local") // Important: Get the time zone
	if err != nil {
		return -1, err
	}
	t, err := time.ParseInLocation(timeLayout, str, loc)
	if err != nil {
		return -1, err
	}
	return t.UnixNano(), nil
}
