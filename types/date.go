package types

import (
	"strings"
	"time"
)

type ZephyrDate struct {
	Date time.Time
}

func (date *ZephyrDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}

	var err error
	date.Date, err = time.Parse("Monday, 2006/01/02", s)
	if err != nil {
		return err
	}

	return nil
}

func (date ZephyrDate) MarshalJSON() ([]byte, error) {
	if date.Date.IsZero() {
		return []byte("\"\""), nil
	}

	fmtDate := date.Date.Format("Monday, 2006/01/02")

	return []byte("\"" + fmtDate + "\""), nil
}

type ZephyrTime struct {
	Time time.Time
}

func (t *ZephyrTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}

	var err error
	t.Time, err = time.Parse("15:04", s)
	if err != nil {
		return err
	}

	return nil
}

func (t ZephyrTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("\"\""), nil
	}

	fmtTime := t.Time.Format("3:04 PM")

	return []byte("\"" + fmtTime + "\""), nil
}

type ZephyrAlertDate struct {
	Date time.Time
}

func (t *ZephyrAlertDate) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "" {
		return nil
	}

	var err error
	t.Date, err = time.Parse("Monday, 2006/01/02 15:04", s)
	if err != nil {
		return err
	}

	return nil
}

func (t ZephyrAlertDate) MarshalJSON() ([]byte, error) {
	if t.Date.IsZero() {
		return []byte("\"\""), nil
	}

	fmtTime := t.Date.Format("Monday, 2006/01/02 3:04 PM")

	return []byte("\"" + fmtTime + "\""), nil
}
