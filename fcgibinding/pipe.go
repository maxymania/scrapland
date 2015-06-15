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
	"io"
	"bufio"
)

type Writer struct{
	*bufio.Reader
	pr *io.PipeWriter
	dest io.Writer
	preoff chan int
	pipeon chan int
	closed chan int
}
func NewWriter(dest io.Writer) *Writer {
	r,w := io.Pipe()
	return &Writer{bufio.NewReader(r),w,dest,make(chan int),make(chan int),make(chan int)}
}
func (this *Writer) Write(p []byte) (n int, err error) {
	select {
	case <- this.preoff:
		select {
		case <- this.pipeon:
		default:
			this.pr.Close()
			<- this.pipeon
		}
		return this.dest.Write(p)
	default:
		return this.pr.Write(p)
	}
	return
}
func (this *Writer) Close() (err error) {
	select {
	case <- this.preoff:
		select {
		case <- this.pipeon:
		default:
			err = this.pr.Close()
		}
	default:
		err = this.pr.Close()
	}
	close(this.closed)
	return
}
func (this *Writer) PipeThrough() {
	close(this.preoff)
	this.WriteTo(this.dest)
	close(this.pipeon)
	<- this.closed
}

