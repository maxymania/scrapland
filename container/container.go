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

package container

import (
	"net/http"
	"text/template"
)

type Element struct{
	ch chan string
}
func NewElement() *Element {
	return &Element{make(chan string,1)}
}
func (e *Element) Get() string {
	s := <- e.ch
	e.ch <- s
	return s
}
func (e *Element) Offer(s string) *Element {
	e.ch <- s
	return e
}

type Page struct{
	Title *Element
	Main *Element
	SideBar []*Element
}

type PageGen interface{
	GetPage (*http.Request)*Page
}

var defaultElement = NewElement().Offer("")

func DefaultElement() *Element { return defaultElement }

type Container struct{
	t *template.Template
	pg PageGen
}

func NewContainer(t *template.Template,pg PageGen) *Container{
	return &Container{t,pg}
}

func (b *Container) ServeHTTP(resp http.ResponseWriter, req *http.Request){
	b.t.Execute(resp,b.pg.GetPage(req))
}

