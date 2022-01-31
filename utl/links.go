// links
package utl

import (
	"log"
	"net/http"

	"golang.org/x/net/html"
)

func Extract(ws *Wlstruct) (*Wlstruct, []Wlstruct, error) {
	var link Wlstruct
	var doc *html.Node
	var resp *http.Response
	if saved, err := SavedFiles.IsSaved(ws.GetLink()); err == nil {
		if saved {
			return nil, nil, nil // first nil is indicator nothing to do
		}
	}

	if ws.IsUdocNil() {
		err := GetContent(ws, Path)
		if err != nil {
			return nil, nil, SPrtErr("links Extract GetContent link=%q  err=%q\n", ws.GetLink(), err)
		}
	}

	if !ws.IsContentTypeHTML() { // parsing no need - exit
		return ws, nil, nil
	}
	doc = ws.GetUdocHTMLNode()
	resp = ws.GetResponse()
	if resp == nil {
		log.Fatal("Internal error: links response is nil")
	}

	var links []Wlstruct
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode {
			if n.Data == "a" || n.Data == "link" {
				for i, a := range n.Attr {
					if a.Key != "href" {
						continue
					}
					lnk, err := resp.Request.URL.Parse(a.Val)
					if err != nil {
						continue // ignore bad URLs
					}
					slnk := lnk.String()
					if Mydomain(slnk, Domain) {
						link.SetDocaddr(&((n.Attr[i]).Val))
						link.SetLink(slnk)
						links = append(links, link)
						l := len(links)
						err := GetContent(&links[l-1], Path)
						if err != nil {
							Prtf("links Extract GetContent2 url=%q  err=%q\n", links[l-1].GetLink(), err)
							links[l-1].SetNeedSave(false)
							links[l-1].SetError()
							continue
						}
					}
				}
			} else if n.Data == "base" { //delete base href
				for i, a := range n.Attr {
					if a.Key != "href" {
						continue
					}
					(n.Attr[i]).Val = Path + "/"
				}

			}
		}
	}
	forEachNode(doc, visitNode, nil)
	doc = nil
	resp = nil
	//----------------------------------------------------------
	return ws, links, nil
}

//!-Extract

// Copied from gopl.io/ch5/outline2.
func forEachNode(n *html.Node, pre, post func(n *html.Node)) {
	if pre != nil {
		pre(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, pre, post)
	}
	if post != nil {
		post(n)
	}
}
