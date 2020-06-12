package main

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
)

const WAPPALYZER_SOURCE_URL = "https://raw.githubusercontent.com/AliasIO/wappalyzer/master/src/apps.json"

type WappalyzerApp struct {
	Name    string
	Website string
	Matches *Matcher
}

func fetchWappalyzerData() (map[string]WappalyzerApp, error) {
	wappalyzerData := map[string]WappalyzerApp{}
	resp, err := sendRequest(WAPPALYZER_SOURCE_URL)
	if err != nil {
		return wappalyzerData, err
	}

	responseBody := make(map[string]map[string]map[string]interface{})

	err = json.Unmarshal(resp.Body, &responseBody)

	for _, value := range responseBody {
		for app, apps := range value {
			match := Matcher{
				Cookies:         nil,
				Icon:            "",
				Headers:         nil,
				ResponseContent: nil,
				Script:          nil,
				JavaScript:      nil,
				Meta:            nil,
			}

			wapp := WappalyzerApp{
				Name:    app,
				Website: "",
				Matches: &match,
			}

			if apps["website"] != nil {
				wapp.Website = apps["website"].(string)
			}

			if apps["icon"] != nil {
				match.Icon = apps["icon"].(string)
			}

			if apps["html"] != nil {
				if err := stringOrSliceHandler(apps["html"], &match.ResponseContent); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer html data", err)
					}
				}
			}

			if apps["headers"] != nil {
				if err := mapHandler(apps["headers"], &match.Headers); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer header data", err)
					}
				}
			}

			if apps["cookies"] != nil {
				if err := mapHandler(apps["cookies"], &match.Cookies); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer cookie data", err)
					}
				}
			}

			if apps["script"] != nil {
				if err := stringOrSliceHandler(apps["script"], &match.Script); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer script data", err)
					}
				}
			}

			if apps["js"] != nil {
				if err := mapHandler(apps["js"], &match.JavaScript); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer js data", err)
					}
				}
			}

			if apps["meta"] != nil {
				if err := mapHandler(apps["meta"], &match.Meta); err != nil {
					if opts.Debug {
						printRed(os.Stderr, "error parsing wappalyzer meta data", err)
					}
				}
			}
			wappalyzerData[strings.ToLower(wapp.Name)] = wapp
		}
	}

	return wappalyzerData, nil
}

func stringOrSliceHandler(value interface{}, matchResult *[]*regexp.Regexp) error {
	errorCount := 0
	matchError := ""

	var matches []*regexp.Regexp

	re, err := stringToRegex(value)
	if err != nil {
		errorCount += 1
		matchError += err.Error() + "\n"
	}
	matches = append(matches, re)

	matches, err = sliceToRegexSlice(value, matches)
	if err != nil {
		errorCount += 1
		matchError += err.Error() + "\n"
	}

	// If both conversions fail, mark as an error and move on
	if errorCount >= 2 {
		return errors.New(matchError)
	} else {
		*matchResult = append(matches)
	}
	return nil
}

func mapHandler(value interface{}, matchResult *map[string]*regexp.Regexp) error {
	headerMap, err := mapToRegexMap(value)
	if err != nil {
		return err
	} else {
		*matchResult = headerMap
	}
	return nil
}
