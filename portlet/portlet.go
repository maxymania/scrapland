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
 The portlet package contains interfaces and structs, that
 allow for a sane and concise implementation of portlets or
 other inlinable things.
 */
package portlet

import(
	"io"
)

// This API is considered unstable and may be extended in the future.
type PortletWriter interface{
	io.Writer
	// SetContentType must be called before any write-method.
	// The main reason for SetContentType is, to offer an hint for any
	// Portlet container, how to handle the content.
	SetContentType(s string)
}

/*
 This interface is inspired by the HttpServletRequest in Java/JEE. Its purpose
 Is to provide an Portlet with informations, that it can use in order to provide
 the right informations at the right time (for example which content to select from
 the database or to scrap from the web, another webapp or another cgi-script).
 
 This API is considered unstable and may be extended in the future.
 */
type PortletRequest interface{
	RawQuery() string
	Parameter(string) string
	Attribute(string) string
}

type Portlet interface{
	ServePortlet(PortletWriter,PortletRequest)
}



