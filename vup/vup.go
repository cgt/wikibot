package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	mwclient "cgt.name/pkg/go-mwclient"
	"cgt.name/pkg/go-mwclient/params"
	"cgt.name/pkg/wikibot"
)

const module = "vup"

// Regexes for changing version fields in "dawiki:Skabelon:Infoboks software"
var (
	// eol: either $, a pipe, or }}
	eol          = `(\s*$|\s*\||\s*}})`
	reStable     = regexp.MustCompile(`(?m:` + `(\s*\|\s*(?:stabil|nyeste_version)\s*=\s*).*` + eol + `)`)
	reStableDate = regexp.MustCompile(`(?m:` + `(\s*\|\s*(?:stabil_dato|nyeste_version_udgivet)\s*=\s*).*` + eol + `)`)
)

func main() {
	log.SetPrefix(module + " ")
	log.SetOutput(os.Stdout)

	mwc, err := wikibot.Setup(module)
	if err != nil {
		log.Printf("error during setup: %v", err)
		os.Exit(1)
	}

	// Prefetch revisions for all pages to be updated in a single request.
	titles := make([]string, 0, len(updaters))
	for _, u := range updaters {
		titles = append(titles, u.PageTitle)
	}
	revisions, err := mwc.GetPagesByName(titles...) // TODO: max n titles?
	if err != nil {
		log.Printf("error fetching wiki revisions: %v", err)
		os.Exit(1)
	}

	// Fetch versions and update pages.
	anyFailed := false
	for _, u := range updaters {
		rev := revisions[u.PageTitle]
		if rev.Error != nil {
			log.Printf("[%s] wiki revision error: %v", u.PageTitle, rev.Error)
			anyFailed = true
			continue
		}
		v, err := u.Fetcher()
		if err != nil {
			log.Printf("[%s] version fetch error: %v", u.PageTitle, rev.Error)
			anyFailed = true
			continue
		}
		newText := reStable.ReplaceAllString(rev.Content, "${1}"+v.Version+"${2}")
		newText = reStableDate.ReplaceAllString(newText, "${1}"+v.FmtDate()+"${2}")

		if newText == rev.Content {
			// No change, skip.
			continue
		}

		err = mwc.Edit(params.Values{
			"title":         u.PageTitle,
			"text":          newText,
			"summary":       v.FmtSummary(),
			"basetimestamp": rev.Timestamp,
			"minor":         "",
			"nocreate":      "",
			"bot":           "",
		})
		if err != nil && err != mwclient.ErrEditNoChange {
			anyFailed = true
			log.Printf("[%s] edit error: %v", u.PageTitle, err)
		}
	}
	if anyFailed {
		os.Exit(1)
	}
}

var updaters = []struct {
	PageTitle string
	Fetcher   VersionFetcher
}{
	{
		"Linux",
		FetchLinux,
	},
}

type VersionFetcher func() (Version, error)

type Version struct {
	Version string
	Date    time.Time
}

// FmtDate formats a Version's Date in the format "2. januar 2006".
// Note the Danish spelling of the month; time.Time.Format uses English.
func (v Version) FmtDate() string {
	var month string

	switch v.Date.Month() {
	case time.January:
		month = "januar"
	case time.February:
		month = "februar"
	case time.March:
		month = "marts"
	case time.April:
		month = "april"
	case time.May:
		month = "maj"
	case time.June:
		month = "juni"
	case time.July:
		month = "juli"
	case time.August:
		month = "august"
	case time.September:
		month = "september"
	case time.October:
		month = "oktober"
	case time.November:
		month = "november"
	case time.December:
		month = "december"
	default:
		panic(fmt.Errorf("invalid month: %v", v.Date.Month()))
	}

	return fmt.Sprintf("%d. %s %d", v.Date.Day(), month, v.Date.Year())
}

// FmtSummary formats an edit summary for a version update edit.
func (v Version) FmtSummary() string {
	return fmt.Sprintf("Opdaterer versionsinformation (%s)", v.Version)
}

// httpc is an *http.Client with a timeout set (unlike http.DefaultClient).
var httpc = &http.Client{
	Timeout: 15 * time.Second,
}
