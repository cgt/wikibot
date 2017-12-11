package main

import (
	"errors"
	"io/ioutil"
	"regexp"
	"time"
)

var (
	gitReVer  = regexp.MustCompile(`<span class="version">\s*([0-9.]*)\s*</span>`)
	gitReDate = regexp.MustCompile(`<span class="release-date">\s*\(([0-9-]*)\)\s*</span>`)
)

func FetchGit() (v Version, err error) {
	resp, err := httpc.Get("http://www.git-scm.com/")
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	soup, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	if gitReVer.Match(soup) {
		matches := gitReVer.FindSubmatch(soup)
		v.Version = string(matches[1])
	} else {
		return v, errors.New("version not found")
	}

	if gitReDate.Match(soup) {
		matches := gitReDate.FindSubmatch(soup)
		t, err := time.Parse("2006-01-02", string(matches[1]))
		if err != nil {
			return v, err
		}
		v.Date = t
	} else {
		return v, errors.New("version date not found")
	}

	return v, nil
}
