package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
)

type Task struct {
	Url string
}

type MatchResult struct {
	Url               string
	TechnologyMatches map[string][]string
}

var config Config
var opts CliOptions
var failedRequestsSent int
var successfulRequestsSent int

var printGreen = color.New(color.FgGreen).PrintfFunc()
var printRed = color.New(color.FgRed).FprintfFunc()
var printCyan = color.New(color.FgCyan).FprintfFunc()

func main() {
	// Create an empty config object
	config = NewConfig()

	// Verify flags are properly formatted/expected
	err := verifyFlags(&opts)
	if err != nil {
		printRed(os.Stderr, "error parsing flags: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Get the URLs provided, deduplicate, and load properly formatted ones into slice
	urls, err := getUrlsFromFile()
	if err != nil {
		fmt.Println("Error getting URLs from stdin: ", err)
	}

	// Create HTTP Transport and Client after parsing flags
	createClient()

	// Fetch the latest wappalyzer data
	config.TechInScope, err = fetchWappalyzerData()
	if err != nil {
		fmt.Println("Error fetching data from Wappalyzer: ", err)
	}

	// Check if specific technology to lookup, else include all
	updateTechnologyInScope()

	tasks := make(chan Task)
	var wg sync.WaitGroup

	for i := 0; i < opts.Concurrency; i++ {
		wg.Add(1)
		go func() {
			for task := range tasks {
				task.execute()
			}
			wg.Done()
		}()
	}

	for _, u := range urls {
		tasks <- Task{Url: u}
	}

	close(tasks)
	wg.Wait()
}

func (t Task) execute() {
	resp, err := sendRequest(t.Url)
	if err != nil {
		failedRequestsSent += 1
		if opts.Debug {
			printRed(os.Stderr, "error sending HTTP request to %v: %v\n", t.Url, err)
		}
		return
	}
	successfulRequestsSent += 1

	responseBody := string(resp.Body)
	if responseBody == "" {
		return
	}

	// Extract relevant data from HTML docs
	htmlExtractions := HtmlExtractions{
		ScriptTags:       []string{},
		InlineJavaScript: []string{},
		MetaTags:         map[string]string{},
	}
	htmlExtractions.getScriptTags(resp.GoQueryDoc)
	htmlExtractions.getInlineJavaScript(resp.GoQueryDoc)
	htmlExtractions.getMetaTags(resp.GoQueryDoc)

	techMatches := map[string][]string{}
	matchResult := MatchResult{
		Url:               t.Url,
		TechnologyMatches: techMatches,
	}

	for key, value := range config.TechInScope {
		var matchTypes []string

		if contentMatch := value.Matches.contentMatch(&responseBody); contentMatch {
			matchTypes = append(matchTypes, "htmlContent")
			matchResult.TechnologyMatches[key] = matchTypes
		}

		if scriptMatch := value.Matches.scriptMatch(&htmlExtractions.ScriptTags); scriptMatch {
			matchTypes = append(matchTypes, "scriptTag")
			matchResult.TechnologyMatches[key] = matchTypes
		}

		if metaMatch := value.Matches.metaMatch(&htmlExtractions.MetaTags); metaMatch {
			matchTypes = append(matchTypes, "metaTag")
			matchResult.TechnologyMatches[key] = matchTypes
		}

		if jsMatch := value.Matches.javascriptMatch(&htmlExtractions.InlineJavaScript); jsMatch {
			matchTypes = append(matchTypes, "javascriptContent")
			matchResult.TechnologyMatches[key] = matchTypes
		}
	}

	for key, value := range config.CustomMatch {
		var matchTypes []string
		matches := []*regexp.Regexp{value}

		if strings.ToLower(key) == "htmlcontent" {
			if match := strAndSliceMatch(&responseBody, matches); match {
				matchTypes = append(matchTypes, "htmlContent")
				matchResult.TechnologyMatches[key] = matchTypes
			}
		}

		if strings.ToLower(key) == "scripttag" {
			if match := sliceAndSliceMatch(&htmlExtractions.ScriptTags, matches); match {
				matchTypes = append(matchTypes, "scriptTag")
				matchResult.TechnologyMatches[key] = matchTypes
			}
		}
	}

	for tech, matchType := range matchResult.TechnologyMatches {
		printGreen("[%v] found [%v] running via the following match types: [%v]\n",
			matchResult.Url, tech, strings.Join(matchType, ", "))
	}
}
