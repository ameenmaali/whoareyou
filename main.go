package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/ameenmaali/whoareyou/pkg/config"
	"github.com/ameenmaali/whoareyou/pkg/matcher"
	"github.com/ameenmaali/whoareyou/pkg/utils"
)

type Task struct {
	Url string
}

var conf config.Config
var opts config.CliOptions
var failedRequestsSent int
var successfulRequestsSent int

func main() {
	// Create an empty conf object
	conf = config.NewConfig()

	// Verify flags are properly formatted/expected
	err := conf.VerifyFlags(&opts)
	if err != nil {
		conf.Utils.PrintRed(os.Stderr, "error parsing flags: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	// Get the URLs provided, deduplicate, and load properly formatted ones into slice
	urls, err := utils.GetUrlsFromFile(&conf)
	if err != nil {
		fmt.Println("Error getting URLs from stdin: ", err)
	}

	// Create HTTP Transport and Client after parsing flags
	conf.HttpClient = utils.CreateClient(opts.Timeout)

	// Fetch the latest wappalyzer data
	conf.TechInScope, err = utils.FetchWappalyzerData(&conf)
	if err != nil {
		fmt.Println("Error fetching data from Wappalyzer: ", err)
	}

	// Check if specific technology to lookup, else include all
	conf.UpdateTechnologyInScope()

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
	resp, err := utils.SendRequest(t.Url, &conf)
	if err != nil {
		failedRequestsSent += 1
		if conf.DebugMode {
			conf.Utils.PrintRed(os.Stderr, "error sending HTTP request to %v: %v\n", t.Url, err)
		}
		return
	}
	successfulRequestsSent += 1

	responseBody := string(resp.Body)
	if responseBody == "" {
		return
	}

	// Extract relevant data from HTML docs
	htmlExtractions := matcher.HtmlExtractions{
		ScriptTags:       []string{},
		InlineJavaScript: []string{},
		MetaTags:         map[string]string{},
	}
	htmlExtractions.Parse(resp.GoQueryDoc)
	htmlExtractions.RawHtmlBody = &responseBody

	techMatches := map[string][]string{}
	matchResult := matcher.MatchResult{
		Url:               t.Url,
		TechnologyMatches: techMatches,
		TechFound:         []string{},
	}

	if !opts.DisableWappalyzer {
		for key, value := range conf.TechInScope {
			value.Matches.HtmlExtractions = htmlExtractions
			value.Matches.Evaluate(key, &matchResult)
		}
	}

	for key, value := range conf.CustomMatch {
		value.Matches.HtmlExtractions = htmlExtractions
		value.Matches.Evaluate(key, &matchResult)
	}

	if len(matchResult.TechFound) > 0 {
		conf.Utils.PrintGreen(os.Stdout, "[%v]: [%v]\n", matchResult.Url, strings.Join(matchResult.TechFound, ", "))
	} else {
		if conf.DebugMode {
			conf.Utils.PrintYellow(os.Stderr, "[%v]: no matches found\n", matchResult.Url)
		}
	}
}
