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
)

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


