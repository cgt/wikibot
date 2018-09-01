package main

import (
	"bytes"
	"fmt"

	"cgt.name/pkg/go-mwclient"
	"cgt.name/pkg/go-mwclient/params"
)

func (w *Bot) generateReport() string {
	// missing templates
	full := difference(w.pfull, w.tfull)
	semi := difference(w.psemi, w.tsemi)

	buf := new(bytes.Buffer)
	buf.WriteString("Denne side indeholder lister over sider, som er beskyttede, men som ikke er påsat en beskyttelsesskabelon. Ikke alle sider nævnt her skal nødvendigvis have påsat en beskyttelsesskabelon.\n")

	buf.WriteString("== Fuldt beskyttede sider, som mangler beskyttelsesskabelon ==\n")
	if len(full) != 0 {
		for _, p := range full {
			fmt.Fprintf(buf, "* [[%s]]\n", p)
		}
	}
	buf.WriteByte('\n')

	if len(semi) != 0 {
		buf.WriteString("== Semi-beskyttede sider, som mangler beskyttelsesskabelon ==\n")
		for _, p := range semi {
			fmt.Fprintf(buf, "* [[%s]]\n", p)
		}
	}
	buf.WriteByte('\n')

	return buf.String()
}

// RemoveOutdated must be run before ReportMissing!
// ReportMissing depends on w.tfull, etc.
func (w *Bot) ReportMissing() error {
	report := w.generateReport()
	err := w.Edit(params.Values{
		"title":    "Bruger:Cgt/Beskyttelsesskabeloner",
		"text":     report,
		"summary":  "Robot: Opdaterer rapport",
		"notminor": "",
		"bot":      "",
	})

	// Suppress unimportant "no change" error
	if err == mwclient.ErrEditNoChange {
		err = nil
	}
	return err
}
