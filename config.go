package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const Version = "1.0.0"

type CliOptions struct {
	Cookies        string
	Headers        string
	Debug          bool
	Concurrency    int
	SilentMode     bool
	Timeout        int
	Version        bool
	RawTechInScope string
}

type Config struct {
	Cookies        string
	Headers        map[string]string
	httpClient     *http.Client
	HasExtraParams bool
	TechProvided   []string
	TechInScope    map[string]WappalyzerApp
}

func verifyFlags(options *CliOptions) error {
	flag.StringVar(&options.Cookies, "cookies", "", "Cookies to add in all requests")

	flag.StringVar(&options.Headers, "H", "", "Headers to add in all requests. Multiple should be separated by semi-colon")
	flag.StringVar(&options.Headers, "headers", "", "Headers to add in all requests. Multiple should be separated by semi-colon")

	flag.StringVar(&options.RawTechInScope, "tech", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")
	flag.StringVar(&options.RawTechInScope, "technology-lookups", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")

	flag.BoolVar(&options.Debug, "debug", false, "Debug/verbose mode to print more info for failed/malformed URLs or requests")

	flag.BoolVar(&options.SilentMode, "s", false, "Only print successful evaluations (i.e. mute status updates). Note these updates print to stderr, and won't be saved if saving stdout to files")
	flag.BoolVar(&options.SilentMode, "silent", false, "Only print successful evaluations (i.e. mute status updates). Note these updates print to stderr, and won't be saved if saving stdout to files")

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
