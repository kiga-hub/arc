package utils

import "time"

//ParseDateTimeString  解析日期时间字符串 2020-05-05 => unix nano
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
	loc, err := time.LoadLocation("Local") //重要：获取时区
	if err != nil {
		return -1, err
	}
	t, err := time.ParseInLocation(timeLayout, str, loc)
	if err != nil {
		return -1, err
	}
	return t.UnixNano(), nil
}
