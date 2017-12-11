package main

import (
	"encoding/json"
	"errors"
	"time"
)

type linux struct {
	LatestStable struct {
		Version string `json:"version"`
	} `json:"latest_stable"`
	Releases []struct {
		Released struct {
			Isodate string `json:"isodate"`
		} `json:"released"`
		Version string `json:"version"`
	} `json:"releases"`
}

func FetchLinux() (v Version, err error) {
	resp, err := httpc.Get("https://www.kernel.org/releases.json")
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	var rlsinfo linux
	json.NewDecoder(resp.Body).Decode(&rlsinfo)
	if err != nil {
		return v, err
	}

	v.Version = rlsinfo.LatestStable.Version
	for _, rls := range rlsinfo.Releases {
		if rls.Version == v.Version {
			t, err := time.Parse("2006-01-02", rls.Released.Isodate)
			if err != nil {
				return v, err
			}
			v.Date = t
			break
		}
	}

	if v.Date.IsZero() {
		return v, errors.New("release date is zero")
	}
	return v, nil
}
