# whoareyou
whoareyou is a tool to find the underlying technology/software used in a list of websites passed through stdin (using Wappalyzer dataset)

## Installation
With Go installed, run:

```
go get -u github.com/ameenmaali/whoareyou
```

## Usage

```
Usage of ./whoareyou:
  -H string
    	Headers to add in all requests. Multiple should be separated by semi-colon
  -V	Get the current version of whoareyou
  -cookies string
    	Cookies to add in all requests
  -debug
    	Debug/verbose mode to print more info for failed/malformed URLs or requests
  -headers string
    	Headers to add in all requests. Multiple should be separated by semi-colon
  -s	Only print successful evaluations (i.e. mute status updates). Note these updates print to stderr, and won't be saved if saving stdout to files
  -silent
    	Only print successful evaluations (i.e. mute status updates). Note these updates print to stderr, and won't be saved if saving stdout to files
  -t int
    	Set the timeout length (in seconds) for each HTTP request (default 15)
  -tech string
    	The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json
  -technology-lookups string
    	The technology to check against (default is all, comma-separated list). Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json
  -timeout int
    	Set the timeout length (in seconds) for each HTTP request (default 15)
  -version
    	Get the current version of whoareyou
  -w int
    	Set the concurrency/worker count (default 25)
  -workers int
    	Set the concurrency/worker count (default 25)
```
