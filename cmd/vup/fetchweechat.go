package main

import (
	"errors"
	"io/ioutil"
	"regexp"
	"time"
)

var reWeeChatVersion = regexp.MustCompile(`<span class="stable">([0-9.]*)</span>`)
var reWeeChatDate = regexp.MustCompile(`<span class="dateversion">(.*)</span>`)

func FetchWeeChat() (v Version, err error) {
	resp, err := httpc.Get("https://weechat.org/")
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()
	soup, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	if m := reWeeChatVersion.FindSubmatch(soup); m != nil {
		v.Version = string(m[1])
	} else {
		return v, errors.New("version not found")
	}

	if m := reWeeChatDate.FindSubmatch(soup); m != nil {
		t, err := time.Parse("Jan 2, 2006", string(m[1]))
		if err != nil {
			return v, err
		}
		v.Date = t
	} else {
		return v, errors.New("version date not found")
	}

	return v, nil
}
