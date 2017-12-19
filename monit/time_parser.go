package monit

import (
	"errors"
	"fmt"
	"time"
)

type TimeParser struct {
	now    time.Time
	str    string
	parsed time.Time
}

func ParseTime(str string) (time.Time, error) {
	now := time.Now()
	return parseTime(now, str)
}

type format struct {
	f string
	h func(time.Time, time.Time) time.Time
}

var formats = []format{
	{
		"2 days ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Day()) * -time.Hour * 24)
		},
	},
	{
		"2 day ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Day()) * -time.Hour * 24)
		},
	},
	{
		"15 hours ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Hour()) * -time.Hour)
		},
	},
	{
		"15 hour ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Hour()) * -time.Hour)
		},
	},
	{
		"4 minutes ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Minute()) * -time.Minute)
		},
	},
	{
		"4 minute ago",
		func(now, t time.Time) time.Time {
			return now.Add(time.Duration(t.Minute()) * -time.Minute)
		},
	},
	{
		"15:04",
		func(now, t time.Time) time.Time {
			return time.Date(now.Year(), now.Month(), now.Day(),
				t.Hour(), t.Minute(),
				0, 0, time.Local)
		},
	},
	{
		"02.01. 15:04",
		func(now, t time.Time) time.Time {
			return time.Date(now.Year(),
				t.Month(), t.Day(), t.Hour(), t.Minute(),
				0, 0, time.Local)
		},
	},
	{
		"2006-01-02T15:04:05",
		func(now, t time.Time) time.Time {
			return t
		},
	},
	{
		"2006-01-02T15:04:05.999999-07:00",
		func(now, t time.Time) time.Time {
			return t
		},
	},
}

var ErrTimeParse = errors.New("time parse")

func parseTime(now time.Time, str string) (time.Time, error) {
	var zero time.Time
	if str == "" {
		return zero, nil
	}

	for _, f := range formats {
		t, err := time.Parse(f.f, str)
		if err != nil {
			continue
		}
		return f.h(now, t), nil
	}

	return zero, ErrTimeParse
}

func TimePatterns() string {
	var r string
	for _, f := range formats {
		r += fmt.Sprintf("\n    %s", f.f)
	}
	return r
}
