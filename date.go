package main

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

const (
	layoutUnix = "Mon Jan 2 15:04:05 -07 2006"
	layoutISO  = "2006-01-02"
)

type publishDate struct {
	time.Time
}

func (d *publishDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var pd string
	if err := unmarshal(&pd); err != nil {
		return err
	}

	layouts := []string{layoutUnix, layoutISO}
	for _, layout := range layouts {
		date, err := time.Parse(layout, pd)
		if err == nil {
			d.Time = date
			return nil
		}
	}

	return errors.Errorf("Unrecognized date format: %s", pd)
}

func toISODate(d publishDate) string {
	return d.Time.Format(layoutISO)
}

func getYear(d publishDate) string {
	return strconv.Itoa(d.Time.Year())
}

func getMonth(d publishDate) string {
	month := int(d.Time.Month())
	if month < 10 {
		return "0" + strconv.Itoa(month)
	}

	return strconv.Itoa(month)
}

func getDay(d publishDate) string {
	day := d.Time.Day()
	if day < 10 {
		return "0" + strconv.Itoa(day)
	}

	return strconv.Itoa(day)
}
