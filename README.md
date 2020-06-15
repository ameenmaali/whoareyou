# whoareyou
whoareyou is a tool to find the underlying technology/software used in a list of URLs 
passed through stdin (using [Wappalyzer](https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json) dataset). It will
make a request to the URL, analyze the data received, and match against known fingerprints/indicators of technology.

Support for custom matches for user provided regex values in HTTP responses is also supported, in addition or standalone from Wappalyzer.

This is useful to understand what technology the website is using, easy search for custom strings/regex, as well as finding many different
websites that use a given set of technology in mass.

## Installation
With Go installed, run:

```
go get -u github.com/ameenmaali/whoareyou
```

## Usage

```
Usage of whoareyou:
  -H string
    	Headers to add in all requests. Multiple should be separated by semi-colon
  -V	Get the current version of whoareyou
  -cookies string
    	Cookies to add in all requests
  -debug
    	Debug/verbose mode to print more info for failed/malformed URLs or requests
  -disable-wappalyzer
    	Disable Wappalyzer scans (useful for only including custom matches)
  -dw
    	Disable Wappalyzer scans (useful for only including custom matches)
  -headers string
    	Headers to add in all requests. Multiple should be separated by semi-colon
  -m value
    	Key value pair (JSON formatted, see README for usage info) of a match source type and regex value (or string) to search for
    	 (i.e. '{"name": {"responseBody": "^http(s)?:\/\/.+"}}'. Available match source types are: responseBody, scriptSrc. Flag can be set more than once.
  -match value
    	Key value pair (JSON formatted, see README for usage info) of a match source type and regex value (or string) to search for
    	 (i.e. '{"name": {"responseBody": "^http(s)?:\/\/.+"}}'. Available match source types are: responseBody, scriptSrc. Flag can be set more than once.
  -tech string
    	The technology to check against (default is all, comma-separated list).
    	 Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json
  -technology-lookups string
    	The technology to check against (default is all, comma-separated list).
    	 Get names from app keys here: https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json
  -t int
    	Set the timeout length (in seconds) for each HTTP request (default 15)
  -timeout int
    	Set the timeout length (in seconds) for each HTTP request (default 15)
  -version
    	Get the current version of whoareyou
  -w int
    	Set the concurrency/worker count (default 25)
  -workers int
    	Set the concurrency/worker count (default 25)
```

### Custom Matches
Support for custom matches is also included with the `-m|-match` flag. This should be a JSON formatted string which
expects a search name (which you create), the match type (where the search should be), and the regex match values you are looking for.

The current supported match types are:
* `responseBody` - Search the entire response body/HTML
* `scriptSrc` - Search for a value within the src tags in scripts in the designated page

Data should be formatted as valid JSON, with the following structure
```
{"searchName": {"matchType": "regexValue"}}
{"searchName": {"matchType": ["regexValue1", "regexValue2"]}}
```

* The `searchName` is whatever you want to identify the search as
* The `matchType` is one of the above supported match types
* The `regexValue`'s as identified should be a string or list of strings (either normal strings or regex values)

You can have as many `-m|-match` flags as you'd like in a given search. To only include custom matches, and not Wappalyzer data,
make sure to include the `-dw|disable-wappalyzer` flag

## Examples

Pass in a list of URLs with no custom matches

```
cat urls.txt | whoareyou
```

Pass in a site to [waybackurls](https://github.com/tomnomnom/waybackurls), run it through [urldedupe](https://github.com/ameenmaali/urldedupe) to deduplicate, and run whoareyou and store to results.txt

```
echo "https://google.com" | waybackurls | urldedupe | whoareyou > results.txt
```

Use a custom match to look for URLs in a response body or script tag

```
cat urls.txt | whoareyou -m '{"findUrls":{"scriptSrc":"^http(s)?:\/\/.+", "responseBody":"^http(s)?:\/\/.+"}}'
```

Use a custom match, and don't use Wappalyzer dataset to look for a specific list of strings in a response body

```
cat urls.txt | whoareyou -m '{"findstring":{"responseBody":["str1","str2","str3"]}}' -dw
```

Search for specify technology key from [Wappalyzer](https://github.com/AliasIO/wappalyzer/blob/master/src/apps.json)

```
cat urls.txt | whoareyou -tech "wordpress,intercom,youtube"
```
