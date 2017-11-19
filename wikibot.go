package wikibot

import (
	"encoding/json"
	"os"

	mwclient "cgt.name/pkg/go-mwclient"
)

const DefaultConfigFile = "$HOME/.wikibot.json"

type Config struct {
	APIURL string `json:"api_url"`
	OAuth  struct {
		ConsumerToken  string `json:"consumer_token"`
		ConsumerSecret string `json:"consumer_secret"`
		AccessToken    string `json:"access_token"`
		AccessSecret   string `json:"access_secret"`
	} `json:"oauth"`
}

func loadConfig() (Config, error) {
	fname := os.ExpandEnv(DefaultConfigFile)
	f, err := os.Open(fname)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func userAgent(module string) string {
	base := "wikibot/1 (chris@cgt.name, Meta:User:Cgt)"
	if module == "" {
		return base
	} else {
		return module + " " + base
	}
}

func Setup(module string) (*mwclient.Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	w, err := mwclient.New(cfg.APIURL, userAgent(module))
	if err != nil {
		return nil, err
	}

	w.Assert = mwclient.AssertBot
	w.Maxlag.On = true

	oa := cfg.OAuth
	err = w.OAuth(oa.ConsumerToken, oa.ConsumerSecret, oa.AccessToken, oa.AccessSecret)
	if err != nil {
		return nil, err
	}

	return w, nil
}
