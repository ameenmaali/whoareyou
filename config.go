package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const Version = "1.0.0"

type CliOptions struct {
	Cookies        string
	Headers        string
	Debug          bool
	Concurrency    int
	Timeout        int
	Version        bool
	RawTechInScope string
	CustomMatch    MultiStringFlag
}

type Config struct {
	Cookies      string
	Headers      map[string]string
	httpClient   *http.Client
	TechProvided []string
	CustomMatch  map[string]*regexp.Regexp
	TechInScope  map[string]WappalyzerApp
}

type MultiStringFlag []string

func NewConfig() Config {
	config := Config{
		Cookies:      "",
		Headers:      make(map[string]string),
		httpClient:   nil,
		TechProvided: []string{},
		CustomMatch:  make(map[string]*regexp.Regexp),
		TechInScope:  make(map[string]WappalyzerApp),
	}
	return config
}

func verifyFlags(options *CliOptions) error {
	flag.StringVar(&options.Cookies, "cookies", "", "Cookies to add in all requests")

	flag.StringVar(&options.Headers, "H", "", "Headers to add in all requests. Multiple should be separated by semi-colon")
	flag.StringVar(&options.Headers, "headers", "", "Headers to add in all requests. Multiple should be separated by semi-colon")

	flag.StringVar(&options.RawTechInScope, "tech", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")
	flag.StringVar(&options.RawTechInScope, "technology-lookups", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")

	flag.Var(&options.CustomMatch, "m", "Key value pair (JSON formatted) of a match source type and regex value (or string) to search for (i.e. '{\"htmlContent\": \"^http(s)?:\\/\\/.+\"}'. Available match source types are: htmlContent, scriptTag")
	flag.Var(&options.CustomMatch, "match", "Key value pair (JSON formatted) of a match source type and regex value (or string) to search for (i.e. '{\"htmlContent\": \"^http(s)?:\\/\\/.+\"}'. Available match source types are: htmlContent, scriptTag")

	flag.BoolVar(&options.Debug, "debug", false, "Debug/verbose mode to print more info for failed/malformed URLs or requests")

	flag.IntVar(&options.Concurrency, "w", 25, "Set the concurrency/worker count")
	flag.IntVar(&options.Concurrency, "workers", 25, "Set the concurrency/worker count")

	flag.IntVar(&options.Timeout, "t", 15, "Set the timeout length (in seconds) for each HTTP request")
	flag.IntVar(&options.Timeout, "timeout", 15, "Set the timeout length (in seconds) for each HTTP request")

	flag.BoolVar(&options.Version, "version", false, "Get the current version of whoareyou")
	flag.BoolVar(&options.Version, "V", false, "Get the current version of whoareyou")

	flag.Parse()

	if options.Version {
		fmt.Println("whoareyou version: " + Version)
		os.Exit(0)
	}

	if options.Cookies != "" {
		config.Cookies = options.Cookies
	}

	if options.Headers != "" {
		if !strings.Contains(options.Headers, ":") {
			return errors.New("headers flag not formatted properly (no colon to separate header and value)")
		}
		headers := make(map[string]string)
		rawHeaders := strings.Split(options.Headers, ";")
		for _, header := range rawHeaders {
			var parts []string
			if strings.Contains(header, ": ") {
				parts = strings.Split(header, ": ")
			} else if strings.Contains(header, ":") {
				parts = strings.Split(header, ":")
			} else {
				continue
			}
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
		config.Headers = headers

	}

	if options.RawTechInScope != "" {
		var technology []string
		rawTechnology := strings.Split(options.RawTechInScope, ",")
		for _, part := range rawTechnology {
			technology = append(technology, strings.ToLower(strings.TrimSpace(part)))
		}
		config.TechProvided = technology
	}

	for _, value := range options.CustomMatch {
		var data map[string]interface{}
		err := json.Unmarshal([]byte(value), &data)
		if err != nil {
			return err
		}

		for key, value := range data {
			str := fmt.Sprintf("%v", value)
			re, err := regexp.Compile(str)
			if err != nil {
				return err
			}
			config.CustomMatch[key] = re
		}
	}

	return nil
}

func updateTechnologyInScope() {
	if config.TechProvided != nil {
		data := map[string]WappalyzerApp{}
		for _, technology := range config.TechProvided {
			if _, ok := config.TechInScope[technology]; ok {
				data[technology] = config.TechInScope[technology]
			} else {
				printRed(os.Stderr, "Technology provided [%v] was not found\n", technology)
			}
		}

		if len(data) != 0 {
			config.TechInScope = data
		}
	}
}

func (m *MultiStringFlag) String() string {
	return ""
}

func (m *MultiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}
