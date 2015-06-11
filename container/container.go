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
 This package implements a simple mechanism, that provides an consistent
 datamodel, that helps, to glue a template together with a content-provider.
 In that turn, it offers the Element object, which enables the leverage of
 concurrency for the site creation with asynchronously generated content.
 */
package container

import (
	"net/http"
	"text/template"
)

/*
 This object models a Future, that contains a string. It is suitable for the
 use in conjunktion with the template engine.
*/
type Element struct{
	ch chan string
}
func NewElement() *Element {
	return &Element{make(chan string,1)}
}

/*
 Returns the string that is contained within the Element object.
 This function will block until the content is offered using the
 Offer() method.
*/
func (e *Element) Get() string {
	s := <- e.ch
	e.ch <- s
	return s
}

/*
 Sets the content within the Element, so every consumer being blocked at
 Get() will unblock.  This method may only called once. Otherwise a deadlock
 occurs.
*/
func (e *Element) Offer(s string) *Element {
	e.ch <- s
	return e
}

type Page struct{
	Title *Element
	Main *Element
	SideBar []*Element
	SiteID string
}
func (p *Page) SiteActive(s string) bool {
	return p.SiteID==s
}

type adtnPage struct{
	*Page
	Additional interface{}
}

type PageGen interface{
	GetPage (*http.Request)*Page
}

var defaultElement = NewElement().Offer("")

func DefaultElement() *Element { return defaultElement }

type Container struct{
	t *template.Template
	pg PageGen
	adtn interface{}
}

func NewContainer(t *template.Template,pg PageGen) *Container{
	return &Container{t,pg,nil}
}
func NewContainerWithAdditional(t *template.Template,pg PageGen,a interface{}) *Container{
	return &Container{t,pg,a}
}

func (b *Container) Additional(a interface{}){ b.adtn = a }

func (b *Container) ServeHTTP(resp http.ResponseWriter, req *http.Request){
	b.t.Execute(resp,&adtnPage{b.pg.GetPage(req),b.adtn})
}

