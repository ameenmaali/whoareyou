package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ameenmaali/whoareyou/pkg/matcher"
	"github.com/fatih/color"
)

const Version = "1.0.0"

type CliOptions struct {
	Cookies           string
	Headers           string
	Debug             bool
	DisableWappalyzer bool
	Concurrency       int
	Timeout           int
	Version           bool
	RawTechInScope    string
	CustomMatch       MultiStringFlag
}

type Config struct {
	Cookies      string
	Headers      map[string]string
	HttpClient   *http.Client
	TechProvided []string
	CustomMatch  map[string]matcher.AppMatch
	TechInScope  map[string]matcher.AppMatch
	Utils        Utilities
	DebugMode bool
}

type PrintColor func(w io.Writer, format string, a ...interface{})

type Utilities struct {
	PrintRed    PrintColor
	PrintGreen  PrintColor
	PrintCyan   PrintColor
	PrintYellow PrintColor
}

type MultiStringFlag []string

func NewConfig() Config {
	utilities := Utilities{
		PrintGreen:  color.New(color.FgGreen).FprintfFunc(),
		PrintRed:    color.New(color.FgRed).FprintfFunc(),
		PrintCyan:   color.New(color.FgCyan).FprintfFunc(),
		PrintYellow: color.New(color.FgYellow).FprintfFunc(),
	}

	config := Config{
		Cookies:      "",
		Headers:      make(map[string]string),
		HttpClient:   nil,
		TechProvided: []string{},
		CustomMatch:  make(map[string]matcher.AppMatch),
		TechInScope:  make(map[string]matcher.AppMatch),
		Utils:        utilities,
	}
	return config
}

func (c *Config) UpdateTechnologyInScope() {
	if c.TechProvided != nil {
		data := map[string]matcher.AppMatch{}
		for _, technology := range c.TechProvided {
			if _, ok := c.TechInScope[technology]; ok {
				data[technology] = c.TechInScope[technology]
			} else {
				c.Utils.PrintRed(os.Stderr, "Technology provided [%v] was not found\n", technology)
			}
		}

		if len(data) != 0 {
			c.TechInScope = data
		}
	}
}

func (c *Config) VerifyFlags(options *CliOptions) error {
	flag.StringVar(&options.Cookies, "cookies", "", "Cookies to add in all requests")

	flag.StringVar(&options.Headers, "H", "", "Headers to add in all requests. Multiple should be separated by semi-colon")
	flag.StringVar(&options.Headers, "headers", "", "Headers to add in all requests. Multiple should be separated by semi-colon")

	flag.StringVar(&options.RawTechInScope, "tech", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")
	flag.StringVar(&options.RawTechInScope, "technology-lookups", "", "The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json")

	flag.Var(&options.CustomMatch, "m", "Key value pair (JSON formatted) of a match source type and regex value (or string) to search for (i.e. '{\"htmlContent\": \"^http(s)?:\\/\\/.+\"}'. Available match source types are: htmlContent, scriptTag")
	flag.Var(&options.CustomMatch, "match", "Key value pair (JSON formatted) of a match source type and regex value (or string) to search for (i.e. '{\"htmlContent\": \"^http(s)?:\\/\\/.+\"}'. Available match source types are: htmlContent, scriptTag")

	flag.BoolVar(&options.DisableWappalyzer, "dw", false, "Disable Wappalyzer scans (useful for only including custom matches)")
	flag.BoolVar(&options.DisableWappalyzer, "disable-wappalyzer", false, "Disable Wappalyzer scans (useful for only including custom matches)")

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
		c.Cookies = options.Cookies
	}

	if options.Debug {
		c.DebugMode = true
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
		c.Headers = headers

	}

	if options.RawTechInScope != "" {
		var technology []string
		rawTechnology := strings.Split(options.RawTechInScope, ",")
		for _, part := range rawTechnology {
			technology = append(technology, strings.ToLower(strings.TrimSpace(part)))
		}
		c.TechProvided = technology
	}

	err := c.parseCustomMatches(options.CustomMatch)
	if err != nil {
		return err
	}

	return nil
}

func (m *MultiStringFlag) String() string {
	return ""
}

func (m *MultiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func (c *Config) parseCustomMatches(data MultiStringFlag) error {
	for _, value := range data {
		var data map[string]map[string]interface{}
		err := json.Unmarshal([]byte(value), &data)
		if err != nil {
			return err
		}

		match := matcher.Matcher{}
		app := matcher.AppMatch{
			Matches: &match,
		}

		for key, value := range data {
			app.Name = "custom-" + key
			for matchType, matchValue := range value {
				var matchValues []*regexp.Regexp
				valType := fmt.Sprintf("%T", matchValue)

				// Not a great way to do this, but...
				if valType == "string" || valType == "float64" {
					str := fmt.Sprintf("%v", matchValue)
					re, err := regexp.Compile(str)
					if err != nil {
						return err
					}
					matchValues = append(matchValues, re)

				} else if valType == "[]interface {}" {
					for _, v := range matchValue.([]interface{}) {
						str := fmt.Sprintf("%v", v)
						re, err := regexp.Compile(str)
						if err != nil {
							return err
						}
						matchValues = append(matchValues, re)
					}
				} else {
					return errors.New(fmt.Sprintf("%v data type is not supported. It must be either a string or list of regex values", matchValue))
				}

				matchType = strings.ToLower(matchType)
				if matchType == "responsebody" {
					match.ResponseContent = matchValues
				} else if matchType == "scriptsrc" {
					match.Script = matchValues
				} else {
					return errors.New(fmt.Sprint("%v is not a valid match type. See the usage info and README for current supported types", matchType))
				}
			}
			c.CustomMatch[app.Name] = app
		}
	}
	return nil
}
