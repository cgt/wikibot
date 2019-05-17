package main

import (
	"testing"
	"time"
)

func TestExtractFirefoxVersion(t *testing.T) {
	v, err := extractFirefoxVersion(releaseNotesPageSnippet)
	if err != nil {
		t.Fatal(err)
	}
	year, month, day := v.Date.Date()
	if year != 2019 || month != time.May || day != 7 {
		t.Errorf("expected 2019-05-07, got %v-%v-%v", year, month, day)
	}
	expectedVersion := "66.0.5"
	if v.Version != expectedVersion {
		t.Errorf("expected version %v, got version %v", expectedVersion, v.Version)
	}
}

const releaseNotesPageSnippet = `
<div class="mzp-l-sidebar">
	<h2 class="c-release-summary">
		<div class="c-release-version">66.0.5</div>
		<div class="c-release-product">Firefox Release</div>
	</h2>
	<p class="c-release-date">May 7, 2019</p> 
</div>
`
