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
 Override implements a http.Handler, that enables prefix matching based routing.
 For example, if you want to serve an static directory using http.FileServer and
 http.Dir, but you want to redirect the path "/" to your dynamic main page, and
 the path "/wiki/<wiki-page>" to your Wiki, and the path "/<yyyy>/<mm>/<dd>/<title>"
 to your blog entrys, you can use "override" to do the job.
 */
package override

import (
	"net/http"
)

type Patterned struct{
	http.Handler
	Prefix string
	Exact string
}
func (p *Patterned) Match(r *http.Request) bool{
	path := r.URL.Path
	if p.Prefix!="" {
		lpre := len(p.Prefix)
		if len(path)<lpre { return false }
		return path[:lpre]==p.Prefix
	} else if p.Exact!="" {
		return path==p.Exact
	}
	return false
}

type Overrider struct{
	Base http.Handler
	Overs []*Patterned
}
func (o *Overrider) Add(prefix string, h http.Handler) {
	o.Overs = append(o.Overs,&Patterned{Handler:h,Prefix:prefix})
}
func (o *Overrider) AddExact(exact string, h http.Handler) {
	o.Overs = append(o.Overs,&Patterned{Handler:h,Exact:exact})
}
func (o *Overrider) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for _,p := range o.Overs {
		if p.Match(req) {
			p.ServeHTTP(resp,req)
			return
		}
	}
	o.Base.ServeHTTP(resp,req)
}



