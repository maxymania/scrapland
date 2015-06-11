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

package tmplhelp

import (
	"net/http"
	"encoding/json"
)

// Loads a template from a http.FileSystem object
func LoadJson(fs http.FileSystem,n string) (interface{},error) {
	var i interface{} = nil
	f,e := fs.Open(n)
	if e!=nil { return nil,e }
	e = json.NewDecoder(f).Decode(&i)
	f.Close()
	return i,e
}
