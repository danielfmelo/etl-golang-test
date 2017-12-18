package utils

import (
	"time"
	"strings"
	"github.com/araddon/dateparse"
)

type Worker struct {
	Name   string
	Job    string
	Salary float64
	Date   time.Time
}

func ConvertToTime(value string) (t time.Time, err error) {

	date := strings.Replace(value, "/", "/25/", -1)
	dateTime, err := dateparse.ParseLocal(date)
	if err != nil {
		return dateTime, err
	}
	return dateTime, nil
}