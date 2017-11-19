package main

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"cgt.name/pkg/go-mwclient/params"
)

var (
	reSemi = regexp.MustCompile(`(\n)??{{(?i:(Skabelon\:)?Semi(beskyttet)?)}}(\n)??`)
	reFull = regexp.MustCompile(`(\n)??{{(?i:(Skabelon\:)?(Skrive)?beskyttet)}}(\n)??`)
)

func strSortStable(a []string) {
	sort.Stable(sort.StringSlice(a))
}

func (w *Bot) RemoveOutdated() error {
	tfull, tsemi, err := w.transcludesProtTmpl()
	if err != nil {
		return err
	}
	pfull, err := w.protectedPages("sysop", nsFromTitles(tfull...)...)
	if err != nil {
		return err
	}
	psemi, err := w.protectedPages("autoconfirmed", nsFromTitles(tsemi...)...)
	if err != nil {
		return err
	}
	strSortStable(tfull)
	strSortStable(tsemi)
	strSortStable(pfull)
	strSortStable(psemi)

	w.tfull = tfull
	w.tsemi = tsemi
	w.pfull = pfull
	w.psemi = psemi

	// unpfull contains the pages that have the full-protection template
	// but are not full-protected.
	unpfull := difference(tfull, pfull)
	unpsemi := difference(tsemi, psemi)

	//fmt.Printf("tfull\t%v\npfull\t%v\nunpfull\t%v\n", tfull, pfull, unpfull)
	//fmt.Printf("tsemi\t%v\npsemi\t%v\nunpsemi\t%v\n", tsemi, psemi, unpsemi)

	if len(unpfull) > 0 {
		if err := w.removeTmpl(unpfull, reFull); err != nil {
			return err
		}
	}
	if len(unpsemi) > 0 {
		if err := w.removeTmpl(unpsemi, reSemi); err != nil {
			return err
		}
	}

	return nil
}

func (w *Bot) transcludesProtTmpl() (full []string, semi []string, err error) {
	const (
		tmplSemi = "Skabelon:Semibeskyttet"
		tmplFull = "Skabelon:Skrivebeskyttet"
	)

	p := params.Values{
		"prop":    "transcludedin",
		"tiprop":  "title",
		"tishow":  "!redirect",
		"tilimit": "max",
	}
	p.AddRange("titles", tmplFull, tmplSemi)

	q := w.NewQuery(p)
	for q.Next() {
		j, err := q.Resp().GetObjectArray("query", "pages")
		if err != nil {
			return nil, nil, err
		}
		for _, tmpl := range j {
			tmplTitle, err := tmpl.GetString("title")
			if err != nil {
				return nil, nil, err
			}

			pages, err := tmpl.GetObjectArray("transcludedin")
			if err != nil {
				return nil, nil, err
			}

			for _, page := range pages {
				title, err := page.GetString("title")
				if err != nil {
					return nil, nil, err
				}
				if tmplTitle == tmplFull {
					full = append(full, title)
				} else if tmplTitle == tmplSemi {
					semi = append(semi, title)
				} else {
					log.Printf("transcludesProtTmpl: unknown template %v", tmplTitle)
				}
			}
		}
	}
	if q.Err() != nil {
		return nil, nil, q.Err()
	}
	return full, semi, nil
}

// protectedPages returns the protected pages with the given protection level
// in the given namespaces.
func (w *Bot) protectedPages(level string, namespaces ...string) ([]string, error) {
	type AllPages struct {
		Query struct {
			AllPages []struct {
				Title string `json:"title"`
			} `json:"allpages"`
		} `json:"query"`
	}

	p := params.Values{
		"list":      "allpages",
		"apprtype":  "edit",
		"aplimit":   "max",
		"apprlevel": level,
	}
	var pages []string

	w.namespaces()
	for _, ns := range namespaces {
		if nsid, ok := w.ns[ns]; ok {
			p["apnamespace"] = nsid
		} else {
			return nil, fmt.Errorf("unknown namespace '%s'", ns)
		}
		q := w.NewQuery(p)
		for q.Next() {
			j, err := q.Resp().MarshalJSON()
			if err != nil {
				return nil, err
			}

			var ap AllPages
			if err := json.Unmarshal(j, &ap); err != nil {
				return nil, err
			}

			for _, p := range ap.Query.AllPages {
				pages = append(pages, p.Title)
			}
		}
		if q.Err() != nil {
			return nil, q.Err()
		}
	}

	return pages, nil
}

func (w *Bot) removeTmpl(pageNames []string, re *regexp.Regexp) error {
	if len(pageNames) == 0 {
		return nil
	}
	pages, err := w.GetPagesByName(pageNames...)
	if err != nil {
		return fmt.Errorf("error getting pages: %v", err)
	}
	for title, p := range pages {
		if p.Error != nil {
			log.Printf("Warning: uncaught page error '%s': %v", title, p.Error)
			continue
		}
		newBody := re.ReplaceAllString(p.Content, "")

		if newBody == p.Content {
			continue
		}
		newBody = strings.TrimSpace(newBody)

		edit := params.Values{
			"title":         title,
			"basetimestamp": p.Timestamp,
			"text":          newBody,
			"summary":       "Robot: Fjerner forældet beskyttelsesskabelon",
			"bot":           "",
			"minor":         "",
			"nocreate":      "",
		}
		err = w.Edit(edit)
		if err != nil {
			return fmt.Errorf("error editing %s: %v", title, err)
		}
		log.Printf("Removed outdated protection template from %s", title)
	}
	return nil
}
