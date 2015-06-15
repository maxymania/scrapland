/*
Copyright 2015 Simon Schmidt
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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


