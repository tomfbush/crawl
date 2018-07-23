package crawler

import (
	"net/http"
	"net/url"

	"github.com/benjaminestes/crawl/src/scrape"
	"golang.org/x/net/html"
)

type Pair struct {
	Key string
	Val string
}

type Result struct {
	Depth       int
	Description string
	Title       string
	H1          string
	Robots      string
	Canonical   string
	Status      string
	StatusCode  int
	Proto       string
	ProtoMajor  int
	ProtoMinor  int
	Header      []*Pair
	Links       []*Link
	Hreflang    []*Hreflang
	*Address
}

func MakeResult(addr *Address, depth int) *Result {
	return &Result{
		Address: addr,
		Depth:   depth,
	}
}

func (r *Result) Hydrate(resp *http.Response, doc *html.Node) {
	for k := range resp.Header {
		r.Header = append(r.Header, &Pair{k, resp.Header.Get(k)})
	}

	// Populate response fields
	r.Status = resp.Status
	r.StatusCode = resp.StatusCode
	r.Proto = resp.Proto
	r.ProtoMajor = resp.ProtoMajor
	r.ProtoMinor = resp.ProtoMinor

	// Populate Content fields
	scrapeResult(r, doc)

	// Populate Hreflang fields
	r.Hreflang = getHreflang(r.URL, doc)

	// Populate and update links
	r.Links = getLinks(r.URL, doc)
}

func scrapeResult(n *Result, doc *html.Node) {
	n.Title = scrape.GetText(scrape.QueryNode("title", nil, doc))
	n.H1 = scrape.GetText(scrape.QueryNode("h1", nil, doc))
	n.Description = scrape.GetAttribute("content", scrape.QueryNode("meta", map[string]string{
		"name": "description",
	}, doc))
	n.Robots = scrape.GetAttribute("content", scrape.QueryNode("meta", map[string]string{
		"name": "robots",
	}, doc))
	n.Canonical = scrape.GetAttribute("href", scrape.QueryNode("link", map[string]string{
		"rel": "canonical",
	}, doc))
}

// FIXME: Should get the same URL resolving treatment as links
func getHreflang(base *url.URL, n *html.Node) (hreflang []*Hreflang) {
	nodes := scrape.QueryNodes("link", map[string]string{
		"rel": "alternate",
	}, n)

	for _, n := range nodes {
		lang := scrape.GetAttribute("hreflang", n)
		href := scrape.GetAttribute("href", n)
		if href != "" {
			hreflang = append(hreflang, MakeHreflang(href, lang))
		}
	}

	return
}

func getLinks(base *url.URL, n *html.Node) (links []*Link) {
	els := scrape.GetNodesByTagName("a", n)
	for _, a := range els {
		h, _ := url.Parse(scrape.GetAttribute("href", a))
		t := base.ResolveReference(h)
		t.Fragment = ""
		link := MakeLink(
			t.String(),
			scrape.GetText(a),
			scrape.GetAttribute("rel", n) == "nofollow", // FIXME: Trim whitespace
		)
		links = append(links, link)
	}
	return links
}
