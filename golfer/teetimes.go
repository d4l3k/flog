package golfer

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TeeTime struct {
	ID         int         `json:"id"`
	CourseID   int         `json:"course_id"`
	StartTime  string      `json:"start_time"`
	Date       string      `json:"date"`
	EventID    interface{} `json:"event_id"`
	Hole       int         `json:"hole"`
	Round      int         `json:"round"`
	Active     bool        `json:"active"`
	Format     string      `json:"format"`
	Blocked    bool        `json:"blocked"`
	Clone      bool        `json:"clone"`
	FreeSlots  int         `json:"free_slots"`
	CartsCount int         `json:"carts_count"`
	CreatedAt  string      `json:"created_at"`
	Departure  interface{} `json:"departure"`
}

const DateFormat = "2006-01-02T15:04"

func (t TeeTime) Time() (time.Time, error) {
	return time.ParseInLocation(DateFormat, fmt.Sprintf("%sT%s", t.Date, t.StartTime), time.Local)
}

func affiliationTypeIDs(af Affiliation, players int) string {
	affiliationTypeID := strconv.Itoa(af.AffiliationTypeID)
	return strings.Join(repeatString(affiliationTypeID, players), ",")
}

func (g *Golfer) TeeTimes(af Affiliation, c Course, date string, players int) ([]TeeTime, error) {
	url := fmt.Sprintf(teetimeAPI, affiliationTypeIDs(af, players), date, c.ID)

	var tt []TeeTime
	if err := g.getJSON(url, &tt); err != nil {
		return nil, err
	}
	return tt, nil
}
