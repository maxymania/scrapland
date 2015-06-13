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

/*
 This is an utility library for use with golang.org/x/net/html.
 Its main purpose is, to extract parts of HTML files and to Serialize them.
 */
package htmlscrape

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"io"
	"regexp"
	"bytes"
)

var whiteSpace = regexp.MustCompile(`\s+`)

type Transf func(*html.Node)

func Chain(t ...Transf) Transf{
	return func(h *html.Node){
		for _,ti := range t { ti(h) }
	}
}

func ReplaceHref(f func(string)string) Transf{
	return func(h *html.Node){
		for i,attr := range h.Attr {
			if attr.Key=="href" { h.Attr[i].Val = f(attr.Val) }
		}
	}
}

func walk(begin, end *html.Node,fn Transf) {
	if begin==nil { return }
	for {
		fn(begin)
		walk(begin.FirstChild,begin.LastChild,fn)
		if end==begin { return }
		begin = begin.NextSibling
	}
}

func Walk(n *html.Node,fn Transf){
	walk(n,n,fn)
}

func find(begin, end *html.Node) (t *html.Node,b *html.Node){
	if begin==nil { return }
	for {
		switch begin.Type {
		case html.ElementNode:
			switch begin.DataAtom {
			case atom.Html,atom.Head:
				t,b = find(begin.FirstChild,begin.LastChild)
			case atom.Body:
				b = begin
			case atom.Title:
				t = begin
			}
		case html.DocumentNode:
			t,b = find(begin.FirstChild,begin.LastChild)
		}
		if end==begin { return }
		begin = begin.NextSibling
	}
	return
}
// finds title and body in an html doc
func FindTB(n *html.Node) (t *html.Node,b *html.Node){ return find(n,n) }

// like html.Render but renders only the child elements.
func Render(w io.Writer,n *html.Node) error{
	if n==nil { return nil }
	begin := n.FirstChild
	end := n.LastChild
	if begin==nil { return nil }
	for {
		e := html.Render(w,begin)
		if e!=nil { return e }
		if end==begin { return nil }
		begin = begin.NextSibling
	}
	return nil
}

func findAttr(h *html.Node,k string) string {
	for _,a := range h.Attr {
		if a.Key==k { return a.Val }
	}
	return ""
}

func lurkFor(begin, end *html.Node,sel string) *html.Node{
	if begin==nil { return nil }
	for {
		switch begin.Type {
		case html.ElementNode:
			switch sel[0]{
			case '.':
				for _,cls := range whiteSpace.Split(findAttr(begin,"class"),-1) {
					if cls==sel[1:] { return begin }
				}
			case '#':
				if findAttr(begin,"id")==sel[1:] { return begin }
			default:
				if begin.Data==sel { return begin }
			}
			fallthrough
		case html.DocumentNode:
			res := lurkFor(begin.FirstChild,begin.LastChild,sel)
			if res!=nil { return res }
		}
		if end==begin { return nil }
		begin = begin.NextSibling
	}
	return nil
}
// finds title and body in an html doc
func LurkFor(n *html.Node,sel string) *html.Node{
	if sel=="" { return n }
	if sel[0]=='?' {
		r := lurkFor(n,n,sel)
		if r==nil { return n }
		return r
	}
	return lurkFor(n,n,sel)
}


func extractText(begin, end *html.Node,d io.Writer) {
	if begin==nil { return }
	for {
		switch begin.Type {
		case html.TextNode:
			d.Write([]byte(begin.Data))
		case html.ElementNode,html.DocumentNode:
			extractText(begin.FirstChild,begin.LastChild,d)
		}
		if end==begin { return }
		begin = begin.NextSibling
	}
	return
}
func ExtractText(h *html.Node) string {
	w := &bytes.Buffer{}
	extractText(h,h,w)
	return w.String()
}
func ExtractTextIO(h *html.Node, w io.Writer) {
	extractText(h,h,w)
}


