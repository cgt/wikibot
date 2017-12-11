package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var reFirefox = regexp.MustCompile(`<h2>Version ([0-9.a-z-]+), first offered to Release channel users on (.+)</h2>`)

func FetchFirefox() (v Version, err error) {
	resp, err := http.Get("https://www.mozilla.org/en-US/firefox/notes/")
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	soup, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}
	if m := reFirefox.FindStringSubmatch(string(soup)); m != nil {
		v.Version = m[1]
		reldate, err := time.Parse("January 2, 2006", m[2])
		if err != nil {
			return v, err
		}
		v.Date = reldate
		return v, nil
	}

	return v, errors.New("version not found")
}
