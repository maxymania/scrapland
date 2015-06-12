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

