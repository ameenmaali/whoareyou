package matcher

import (
	"github.com/PuerkitoBio/goquery"
)

type HtmlExtractions struct {
	ScriptTags       []string
	InlineJavaScript []string
	MetaTags         map[string]string
	RawHtmlBody      *string
}

func (he *HtmlExtractions) getScriptTags(doc *goquery.Document) {
	var scripts []string
	doc.Find("script").Each(func(i int, item *goquery.Selection) {
		if src, exists := item.Attr("src"); exists {
			scripts = append(scripts, src)
		}
	})
	he.ScriptTags = scripts
}

func (he *HtmlExtractions) getMetaTags(doc *goquery.Document) {
	doc.Find("meta").Each(func(i int, item *goquery.Selection) {
		attr := item.Get(0)
		for _, a := range attr.Attr {
			he.MetaTags[a.Key] = a.Val
		}
	})
}

func (he *HtmlExtractions) getInlineJavaScript(doc *goquery.Document) {
	var inlineJS []string
	doc.Find("script").Each(func(i int, item *goquery.Selection) {
		inlineJS = append(inlineJS, item.Text())
	})
	he.InlineJavaScript = inlineJS
}

func (he *HtmlExtractions) Parse(doc *goquery.Document) {
	he.getScriptTags(doc)
	he.getMetaTags(doc)
	he.getInlineJavaScript(doc)
}
