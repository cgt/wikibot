package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"cgt.name/pkg/go-mwclient"
	"cgt.name/pkg/go-mwclient/params"
	"cgt.name/pkg/wikibot"
)

const module = "fixprotect"

func main() {
	log.SetPrefix(module + " ")
	log.SetOutput(os.Stdout)

	mwc, err := wikibot.Setup(module)
	if err != nil {
		log.Printf("error during setup: %v", err)
		os.Exit(1)
	}
	w := Bot{mwc, nil, nil, nil, nil, nil}

	if err := w.RemoveOutdated(); err != nil {
		log.Printf("error removing outdated templates: %v", err)
		os.Exit(1)
	}

	if err := w.ReportMissing(); err != nil {
		log.Printf("error creating report on missing templates: %v", err)
		os.Exit(1)
	}
}

type Bot struct {
	*mwclient.Client
	ns           map[string]string
	tfull, tsemi []string
	pfull, psemi []string
}

func (w *Bot) namespaces() error {
	if w.ns != nil {
		return nil
	}

	j, err := w.Get(params.Values{
		"action": "query",
		"meta":   "siteinfo",
		"siprop": "namespaces",
	})
	if err != nil {
		return err
	}
	j, err = j.GetObject("query", "namespaces")
	if err != nil {
		return err
	}

	ns := make(map[string]string)
	for _, v := range j.Map() {
		o, err := v.Object()
		if err != nil {
			return err
		}
		id, err := o.GetNumber("id")
		if err != nil {
			return err
		}
		name, err := o.GetString("name")
		if err != nil {
			return err
		}
		ns[name] = id.String()
	}

	w.ns = ns
	return nil
}

func nsFromTitles(titles ...string) []string {
	var namespaces []string
	for _, t := range titles {
		s := strings.SplitN(t, ":", 2)

		var ns string
		if len(s) == 1 {
			ns = "" // article namespace
		} else if len(s) == 2 {
			ns = s[0]
		} else {
			panic(fmt.Sprintf("nsFromTitles: len(s) == %d: %#v", len(s), s))
		}

		if !contains(namespaces, ns) {
			namespaces = append(namespaces, ns)
		}
	}
	return namespaces
}

// contains searches for a string s in a slice l and returns true if s is found.
func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// difference returns the set difference of a and b (a \ b).
// a and b must be sorted.
func difference(a, b []string) []string {
	var c []string
	for _, x := range a {
		i := sort.SearchStrings(b, x)
		if i == len(b) || i < len(b) && b[i] != x { // not found
			c = append(c, x)
		}
	}
	return c

}
