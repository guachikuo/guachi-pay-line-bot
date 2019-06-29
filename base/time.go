package base

import (
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// ParseToTimestamp will parse timestampStr to unix timestamp
// timestampStr's format: 2019/05/20
func ParseToTimestamp(timestampStr string) (int64, error) {
	const shortForm = "2006/01/02"
	result, err := time.Parse(shortForm, timestampStr)
	if err != nil {
		logrus.WithField("err", err).Error("time.Parse failed in parseToTimestamp")
		return int64(0), err
	}

	location, _ := time.LoadLocation("Asia/Taipei")
	return result.In(location).Unix(), nil
}

// ParseToyyymmdd will parse timestamp to timestampStr
// format: 2019/05/20
func ParseToyyymmdd(timestamp int64) string {
	temp := time.Unix(timestamp, 0)
	// change to Asia/Taipei zone
	location, _ := time.LoadLocation("Asia/Taipei")
	temp = temp.In(location)

	monthStr := strconv.FormatInt(int64(temp.Month()), 10)
	if temp.Month() < time.October {
		monthStr = "0" + monthStr
	}

	dayStr := strconv.FormatInt(int64(temp.Day()), 10)
	if temp.Day() < 10 {
		dayStr = "0" + dayStr
	}
	return fmt.Sprintf("%d/%s/%s", temp.Year(), monthStr, dayStr)
}

// ParseToyyymmddhhmm will parse timestamp to timestampStr
// format: 2019/05/20 12:00
func ParseToyyymmddhhmm(timestamp int64) string {
	temp := time.Unix(timestamp, 0)
	// change to Asia/Taipei zone
	location, _ := time.LoadLocation("Asia/Taipei")
	temp = temp.In(location)

	monthStr := strconv.FormatInt(int64(temp.Month()), 10)
	if temp.Month() < time.October {
		monthStr = "0" + monthStr
	}

	dayStr := strconv.FormatInt(int64(temp.Day()), 10)
	if temp.Day() < 10 {
		dayStr = "0" + dayStr
	}

	hourStr := strconv.FormatInt(int64(temp.Hour()), 10)
	if temp.Hour() < 10 {
		hourStr = "0" + hourStr
	}

	minuteStr := strconv.FormatInt(int64(temp.Minute()), 10)
	if temp.Minute() < 10 {
		minuteStr = "0" + minuteStr
	}
	return fmt.Sprintf("%d/%s/%s %s:%s", temp.Year(), monthStr, dayStr, hourStr, minuteStr)
}
