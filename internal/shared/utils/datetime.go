package utils

import "time"

func ParseDateStringYYYYMMDD(dateStr string) (*time.Time, error) {
	layout := "2006-01-02"
	parsedTime, err := time.Parse(layout, dateStr)
	if err != nil {
		return nil, err
	}
	return &parsedTime, nil
}
