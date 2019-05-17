package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

var reFirefoxVersion = regexp.MustCompile(`<div class="c-release-version">([0-9.a-z-]+)</div>`)
var reFirefoxDate = regexp.MustCompile(`<p class="c-release-date">(.+)</p>`)

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
	v, err = extractFirefoxVersion(string(soup))
	if err != nil {
		err = fmt.Errorf("version not found: %v", err)
	}
	return v, err
}

func extractFirefoxVersion(soup string) (Version, error) {
	var v Version
	version := reFirefoxVersion.FindStringSubmatch(soup)
	if version == nil {
		return v, errors.New("unable to find version in page")
	}
	v.Version = version[1]

	dateText := reFirefoxDate.FindStringSubmatch(soup)
	if dateText == nil {
		return v, errors.New("unable to find release date in page")
	}
	date, err := time.Parse("January 2, 2006", dateText[1])
	if err != nil {
		return v, err
	}
	v.Date = date

	return v, nil
}
