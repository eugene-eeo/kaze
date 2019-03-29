package main

import "regexp"
import "strings"
import "golang.org/x/net/html"

var layoutWhitespace = regexp.MustCompile(`[\n]+`)
var nonlayoutWhitespace = regexp.MustCompile(`[ \t\f]+`)

type Hyperlink struct {
	Text string
	Href string
}

func getAHref(h *html.Node) string {
	for _, attr := range h.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}

func trim(s string) string {
	return strings.TrimSpace(
		nonlayoutWhitespace.ReplaceAllLiteralString(
			layoutWhitespace.ReplaceAllLiteralString(s, "\n"),
			" ",
		),
	)
}

func TextInfoFromString(s string) (string, []Hyperlink) {
	node, err := html.Parse(strings.NewReader(s))
	if err != nil {
		// If we fail to parse then just return the text as-is
		return s, nil
	}
	chunks := []string{}
	links := []Hyperlink{}
	// Otherwise visit
	var f func(*html.Node)
	f = func(node *html.Node) {
		switch node.Type {
		case html.ElementNode:
			switch node.Data {
			case "a":
				if href := getAHref(node); len(href) != 0 {
					// we need to remember the current position of the text chunks
					i := len(chunks)
					for c := node.FirstChild; c != nil; c = c.NextSibling {
						f(c)
					}
					// Find out the text label to give our hyperlink
					j := len(chunks)
					label := strings.Join(chunks[i:j], " ")
					links = append(links, Hyperlink{label, href})
					return
				}
			case "img":
				return
			}
		case html.TextNode:
			chunks = append(chunks, trim(node.Data))
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
	return strings.Join(chunks, " "), links
}
