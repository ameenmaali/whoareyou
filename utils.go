package main

import (
	"bufio"
	"errors"
	"net/url"
	"os"
	"regexp"
	"strings"
)

func getUrlsFromFile() ([]string, error) {
	deduplicatedUrls := make(map[string]bool)
	var urls []string

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		providedUrl := scanner.Text()

		// Only include properly formatted URLs
		u, err := url.ParseRequestURI(providedUrl)
		if err != nil {
			if opts.Debug {
				printRed(os.Stderr, "url provided [%v] is not a properly formatted URL\n", providedUrl)
			}
			continue
		}

		if deduplicatedUrls[u.String()] {
			continue
		}

		deduplicatedUrls[u.String()] = true
		urls = append(urls, u.String())
	}

	return urls, scanner.Err()
}

func stringToRegex(value interface{}) (*regexp.Regexp, error) {
	str, err := cleanString(value)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(str)
	if err != nil {
		return nil, err
	}

	return re, nil
}

func sliceToRegexSlice(value interface{}, matches []*regexp.Regexp) ([]*regexp.Regexp, error) {
	values, ok := value.([]interface{})
	if !ok {
		return matches, errors.New("value provided is not a slice of strings")
	}

	for _, str := range values {
		s, err := cleanString(str)
		if err != nil {
			continue
		}

		re, err := regexp.Compile(s)
		if err != nil {
			continue
		}
		matches = append(matches, re)
	}

	return matches, nil
}

func mapToRegexMap(value interface{}) (map[string]*regexp.Regexp, error) {
	values, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.New("value provided is not a properly formated map")
	}

	regexMap := map[string]*regexp.Regexp{}
	for key, val := range values {
		re, err := regexp.Compile(val.(string))
		if err != nil {
			continue
		}
		regexMap[key] = re
	}
	return regexMap, nil
}

func cleanString(value interface{}) (string, error) {
	str, ok := value.(string)
	if !ok {
		return "", errors.New("value provided is not a string")
	}

	splitStr := strings.Split(str, ";")
	// Only take the first portion of the string, which contains the regex value
	str = splitStr[0]
	if endsWithSlash := strings.HasSuffix(str, "\\"); endsWithSlash {
		str = strings.TrimSuffix(str, "\\")
	}
	return str, nil
}
