package webscrape

import (
	"net/http"
	"github.com/maxymania/scrapland/container"
	"github.com/maxymania/scrapland/htmlscrape"
	"golang.org/x/net/html"
	"bytes"
)

type HttpClient interface{
	Do(req *http.Request) (resp *http.Response, err error)
}


type QueryElement struct{
	Selectors []string
	Element   *container.Element

	// If not empty, Tag contains the tag of the element, which should
	// enclose the content. In that turn, all other HTML tags are stripped,
	// leaving only the bare tags.
	// If Tag is "-", it behaves much like Tag=="", except that the surrounding,
	// html-tag is also present in the output.
	Tag      string
	data      string
}

func GetFragments(hc HttpClient, r *http.Request, qs []*QueryElement) {
	defer func(){
		buf := &bytes.Buffer{}
		for _,q := range qs {
			buf.WriteString(q.data)
			if q.Element!=nil {
				q.Element.Offer(buf.String())
				buf.Reset()
			}
			//q.Element.Offer(q.data)
		}
	}()
	resp,e := hc.Do(r)
	if e!=nil { return }
	p,e := html.Parse(resp.Body)
	if e!=nil { return }
	for _,q := range qs {
		qe := p
		for _,s := range q.Selectors {
			qe = htmlscrape.LurkFor(qe,s)
		}
		if qe==nil { continue }
		if q.Tag=="" {
			buf := &bytes.Buffer{}
			htmlscrape.Render(buf,qe)
			q.data = buf.String()
		} else if q.Tag=="-" {
			buf := &bytes.Buffer{}
			html.Render(buf,qe)
			q.data = buf.String()
		} else {
			t := htmlscrape.ExtractText(qe)
			q.data = "<"+q.Tag+">"+html.EscapeString(t)+"</"+q.Tag+">"
		}
	}
}


