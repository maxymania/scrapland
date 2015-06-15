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

package fcgibinding


import (
	"github.com/maxymania/scrapland/fcgiclient"
	"net/http"
	"fmt"
	"strings"
	"unicode"
)


func MakePool(n int) chan *fcgiclient.FCGIClient {
	cp := make(chan *fcgiclient.FCGIClient,n)
	for i:=0; i<n; i++ { cp <- nil }
	return cp
}

func get(
	p chan *fcgiclient.FCGIClient,
	host string, port interface{}) (*fcgiclient.FCGIClient,error) {
	c := <- p
	if c==nil {
		return fcgiclient.New(host,port)
	}
	if c.Broken() {
		return fcgiclient.New(host,port)
	}
	return c,nil
}

type Handler struct{
	Root string // root URI prefix of handler or empty for "/"
	ServerSoftware string // the server software identifier

	Handlers chan *fcgiclient.FCGIClient
	Host string
	Port interface{}
}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	f,e := get(h.Handlers,h.Host,h.Port)
	h.Handlers <- f
	if e!=nil {
		resp.WriteHeader(500)
		resp.Write([]byte("<h3>Error: 505</h3><p>"))
		fmt.Fprintln(resp,e)
		resp.Write([]byte("</p>"))
		return
	}
	pathInfo := req.URL.Path
	root := h.Root
	if root == "" { root = "/" }
	server := h.ServerSoftware
	if server == "" { server = "go/FastCGI" }
	env := make(map[string]string)
	env["REQUEST_METHOD"] = req.Method
	env["SCRIPT_FILENAME"] = root+pathInfo
	env["SCRIPT_NAME"] = pathInfo
	env["SERVER_SOFTWARE"] = server
	env["SERVER_PROTOCOL"] = "HTTP/1.1"
	env["SERVER_NAME"] = req.Host
	env["SERVER_PORT"] = "80"
	env["QUERY_STRING"] = req.URL.RawQuery
	env["HTTP_HOST"] = req.Host
	env["REQUEST_URI"] = req.URL.RequestURI()
	env["PATH_INFO"] = pathInfo
	env["REMOTE_ADDR"] = req.RemoteAddr
	env["REMOTE_HOST"] = req.RemoteAddr
	w := NewWriter(resp)
	go func(){
		f.RequestIO(env,"",w,nil)
		w.Close()
	}()
	rh := resp.Header()
	for {
		h,_ := w.ReadString('\n')
		h = strings.TrimFunc(h,unicode.IsSpace)
		if h=="" { break }
		hdr := strings.SplitN(h,": ",2)
		if len(hdr)==2 { rh.Add(hdr[0],hdr[1]) }
	}
	w.PipeThrough()
}

