package main

import (
	"testing"
	"time"
)

func TestDateIsBookable(t *testing.T) {
	now = func() time.Time {
		return time.Date(2018, 05, 10, 0, 0, 0, 0, time.Now().Location())
	}
	cases := []struct {
		date string
		want bool
	}{
		{"2018-05-17T07:10", true},
		{"2018-05-18T07:10", true},
		{"2018-05-19T07:10", false},
	}

	for i, c := range cases {
		out, err := dateIsBookable(c.date)
		if err != nil {
			t.Fatal(err)
		}
		if out != c.want {
			t.Errorf("%d. dateIsBookable(%q) = %v; not %v", i, c.date, out, c.want)
		}
	}
}
