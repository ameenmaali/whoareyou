package matcher

import (
	"regexp"
	"strings"
)

type Matcher struct {
	Cookies         map[string]*regexp.Regexp
	Headers         map[string]*regexp.Regexp
	Icon            string
	ResponseContent []*regexp.Regexp
	Script          []*regexp.Regexp
	JavaScript      map[string]*regexp.Regexp
	Meta            map[string]*regexp.Regexp
	HtmlExtractions HtmlExtractions
}

type AppMatch struct {
	Name    string
	Website string
	Matches *Matcher
}

type MatchResult struct {
	Url               string
	TechnologyMatches map[string][]string
	TechFound         []string
}

func (m *Matcher) contentMatch(body *string) bool {
	return strAndSliceMatch(body, m.ResponseContent)
}

func (m *Matcher) headersMatch(header *string) bool {
	return strAndMapMatch(header, m.Headers)
}

func (m *Matcher) cookiesMatch(cookie *string) bool {
	return strAndMapMatch(cookie, m.Cookies)
}

func (m *Matcher) javascriptMatch(js *[]string) bool {
	return sliceAndMapMatch(js, m.JavaScript)
}

func (m *Matcher) scriptMatch(script *[]string) bool {
	return sliceAndSliceMatch(script, m.Script)
}

func (m *Matcher) metaMatch(meta *map[string]string) bool {
	return mapAndMapMatch(meta, m.Meta)
}

func (m *Matcher) Evaluate(tech string, matchResult *MatchResult) {
	var matchTypes []string
	if contentMatch := m.contentMatch(m.HtmlExtractions.RawHtmlBody); contentMatch {
		matchTypes = append(matchTypes, "htmlContent")
		matchResult.TechnologyMatches[tech] = matchTypes
		matchResult.TechFound = append(matchResult.TechFound, tech)
	}

	if scriptMatch := m.scriptMatch(&m.HtmlExtractions.ScriptTags); scriptMatch {
		matchTypes = append(matchTypes, "scriptTag")
		matchResult.TechnologyMatches[tech] = matchTypes
		matchResult.TechFound = append(matchResult.TechFound, tech)
	}

	if metaMatch := m.metaMatch(&m.HtmlExtractions.MetaTags); metaMatch {
		matchTypes = append(matchTypes, "metaTag")
		matchResult.TechnologyMatches[tech] = matchTypes
		matchResult.TechFound = append(matchResult.TechFound, tech)
	}

	if jsMatch := m.javascriptMatch(&m.HtmlExtractions.InlineJavaScript); jsMatch {
		matchTypes = append(matchTypes, "javascriptContent")
		matchResult.TechnologyMatches[tech] = matchTypes
		matchResult.TechFound = append(matchResult.TechFound, tech)
	}
}

func strAndMapMatch(matchStrPtr *string, values map[string]*regexp.Regexp) bool {
	matchStr := *matchStrPtr
	for key, match := range values {
		if match == nil {
			continue
		}

		if strings.ToLower(matchStr) == strings.ToLower(key) && match.MatchString(matchStr) {
			return true
		}
	}
	return false
}

func strAndSliceMatch(matchStrPtr *string, values []*regexp.Regexp) bool {
	matchStr := *matchStrPtr
	for _, match := range values {
		if match == nil {
			continue
		}

		if match.MatchString(matchStr) {
			return true
		}
	}
	return false
}

func sliceAndSliceMatch(matchSlicePtr *[]string, values []*regexp.Regexp) bool {
	matchSlice := *matchSlicePtr
	for _, match := range values {
		if match == nil {
			continue
		}

		for _, val := range matchSlice {
			if match.MatchString(val) {
				return true
			}
		}
	}
	return false
}

func sliceAndMapMatch(matchSlicePtr *[]string, values map[string]*regexp.Regexp) bool {
	matchSlice := *matchSlicePtr
	for key, match := range values {
		if match == nil {
			continue
		}

		for _, val := range matchSlice {
			if strings.ToLower(val) == strings.ToLower(key) && match.MatchString(val) {
				return true
			}
		}
	}
	return false
}

func mapAndMapMatch(matchMapPtr *map[string]string, values map[string]*regexp.Regexp) bool {
	matchMap := *matchMapPtr
	for key, match := range values {
		for attr, val := range matchMap {
			if strings.ToLower(key) == strings.ToLower(attr) && match.MatchString(val) {
				return true
			}
		}
	}
	return false
}
