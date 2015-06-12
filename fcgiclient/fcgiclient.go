// Copyright 2015 Simon Schmidt
// Copyright 2012 Junqing Tan <ivan@mysqlab.net> and The Go Authors
// Use of this source code is governed by a BSD-style
// Part of source code is from Go fcgi package

// Fix bug: Can't recive more than 1 record untill FCGI_END_REQUEST 2012-09-15
// By: wofeiwo

package fcgiclient

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
)

var ConnectionBrokenError = errors.New("fcgi ConnectionBrokenError")

const FCGI_LISTENSOCK_FILENO uint8 = 0
const FCGI_HEADER_LEN uint8 = 8
const VERSION_1 uint8 = 1
const FCGI_NULL_REQUEST_ID uint8 = 0
const FCGI_KEEP_CONN uint8 = 1

const (
	FCGI_BEGIN_REQUEST uint8 = iota + 1
	FCGI_ABORT_REQUEST
	FCGI_END_REQUEST
	FCGI_PARAMS
	FCGI_STDIN
	FCGI_STDOUT
	FCGI_STDERR
	FCGI_DATA
	FCGI_GET_VALUES
	FCGI_GET_VALUES_RESULT
	FCGI_UNKNOWN_TYPE
	FCGI_MAXTYPE = FCGI_UNKNOWN_TYPE
)

const (
	FCGI_RESPONDER uint8 = iota + 1
	FCGI_AUTHORIZER
	FCGI_FILTER
)

const (
	FCGI_REQUEST_COMPLETE uint8 = iota
	FCGI_CANT_MPX_CONN
	FCGI_OVERLOADED
	FCGI_UNKNOWN_ROLE
)

const (
	FCGI_MAX_CONNS  string = "MAX_CONNS"
	FCGI_MAX_REQS   string = "MAX_REQS"
	FCGI_MPXS_CONNS string = "MPXS_CONNS"
)

const (
	maxWrite = 6553500 // maximum record body
	maxPad   = 255
)

type header struct {
	Version       uint8
	Type          uint8
	Id            uint16
	ContentLength uint16
	PaddingLength uint8
	Reserved      uint8
}

// for padding so we don't have to allocate all the time
// not synchronized because we don't care what the contents are
var pad [maxPad]byte

func (h *header) init(recType uint8, reqId uint16, contentLength int) {
	h.Version = 1
	h.Type = recType
	h.Id = reqId
	h.ContentLength = uint16(contentLength)
	h.PaddingLength = uint8(-contentLength & 7)
}

type record struct {
	h   header
	buf [maxWrite + maxPad]byte
}

func (rec *record) read(r io.Reader) (err error) {
	if err = binary.Read(r, binary.BigEndian, &rec.h); err != nil {
		return err
	}
	if rec.h.Version != 1 {
		return errors.New("fcgi: invalid header version")
	}
	n := int(rec.h.ContentLength) + int(rec.h.PaddingLength)
	if _, err = io.ReadFull(r, rec.buf[:n]); err != nil {
		return err
	}
	return nil
}

func (r *record) content() []byte {
	return r.buf[:r.h.ContentLength]
}

type respObj struct{
	id uint16
	done chan struct{}
	out io.Writer
	err io.Writer
}

type FCGIClient struct {
	mutex     sync.Mutex
	rwc       io.ReadWriteCloser
	h         header
	buf       bytes.Buffer
	keepAlive bool
	active    map[uint16]respObj
	stream    chan respObj
	ctr       chan uint16
	broken    chan struct{}
}

/*
 Creates a new FCGI client.

 If the second parameter is an int, the connection to net.Dial("tcp",h+":"+args) is established.
 If the second parameter is an string, the connection to net.Dial("unix",args) is established.
 */
func New(h string, args interface{}) (fcgi *FCGIClient, err error) {
	var conn net.Conn
	switch args.(type) {
	case int:
		addr := h + ":" + strconv.FormatInt(int64(args.(int)), 10)
		conn, err = net.Dial("tcp", addr)
	case string:
		addr := args.(string)
		conn, err = net.Dial("unix", addr)
	default:
		err = errors.New("fcgi: we only accept int (port) or string (socket) params.")
	}
	fcgi = &FCGIClient{
		rwc:       conn,
		keepAlive: true,
		active:    make(map[uint16]respObj),
		stream:    make(chan respObj,1024),
		ctr:       make(chan uint16,1),
		broken:    make(chan struct{}),
	}
	fcgi.ctr <- 1
	go fcgi.worker()
	return
}
func (this *FCGIClient) Broken() bool {
	select {
	case <- this.broken: return true
	default:
	}
	return false
}
func (this *FCGIClient) worker(){
	rec := &record{}
	var err1 error

	// recive forever
	for {
		err1 = rec.read(this.rwc)
		if err1 != nil { this.Close(); break }
		for {
			select {
			case ro := <- this.stream:
				this.active[ro.id]=ro
				continue
			default:
			}
			break
		}
		switch {
		case rec.h.Type == FCGI_STDOUT:
			if ro,ok := this.active[rec.h.Id]; ok {
				if ro.out!=nil { ro.out.Write(rec.content()) }
			}
		case rec.h.Type == FCGI_STDERR:
			if ro,ok := this.active[rec.h.Id]; ok {
				if ro.err!=nil { ro.err.Write(rec.content()) }
			}
		case rec.h.Type == FCGI_END_REQUEST:
			if ro,ok := this.active[rec.h.Id]; ok {
				close(ro.done)
			}
			delete(this.active,rec.h.Id)
		}
	}
}


func (this *FCGIClient) writeRecord(recType uint8, reqId uint16, content []byte) (err error) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.buf.Reset()
	this.h.init(recType, reqId, len(content))
	if err := binary.Write(&this.buf, binary.BigEndian, this.h); err != nil {
		return err
	}
	if _, err := this.buf.Write(content); err != nil {
		return err
	}
	if _, err := this.buf.Write(pad[:this.h.PaddingLength]); err != nil {
		return err
	}
	_, err = this.rwc.Write(this.buf.Bytes())
	return err
}

func (this *FCGIClient) writeBeginRequest(reqId uint16, role uint16, flags uint8) error {
	b := [8]byte{byte(role >> 8), byte(role), flags}
	return this.writeRecord(FCGI_BEGIN_REQUEST, reqId, b[:])
}

func (this *FCGIClient) writeEndRequest(reqId uint16, appStatus int, protocolStatus uint8) error {
	b := make([]byte, 8)
	binary.BigEndian.PutUint32(b, uint32(appStatus))
	b[4] = protocolStatus
	return this.writeRecord(FCGI_END_REQUEST, reqId, b)
}

func (this *FCGIClient) writePairs(recType uint8, reqId uint16, pairs map[string]string) error {
	w := newWriter(this, recType, reqId)
	b := make([]byte, 8)
	for k, v := range pairs {
		n := encodeSize(b, uint32(len(k)))
		n += encodeSize(b[n:], uint32(len(v)))
		if _, err := w.Write(b[:n]); err != nil {
			return err
		}
		if _, err := w.WriteString(k); err != nil {
			return err
		}
		if _, err := w.WriteString(v); err != nil {
			return err
		}
	}
	w.Close()
	return nil
}

func readSize(s []byte) (uint32, int) {
	if len(s) == 0 {
		return 0, 0
	}
	size, n := uint32(s[0]), 1
	if size&(1<<7) != 0 {
		if len(s) < 4 {
			return 0, 0
		}
		n = 4
		size = binary.BigEndian.Uint32(s)
		size &^= 1 << 31
	}
	return size, n
}

func readString(s []byte, size uint32) string {
	if size > uint32(len(s)) {
		return ""
	}
	return string(s[:size])
}

func encodeSize(b []byte, size uint32) int {
	if size > 127 {
		size |= 1 << 31
		binary.BigEndian.PutUint32(b, size)
		return 4
	}
	b[0] = byte(size)
	return 1
}

// bufWriter encapsulates bufio.Writer but also closes the underlying stream when
// Closed.
type bufWriter struct {
	closer io.Closer
	*bufio.Writer
}

func (w *bufWriter) Close() error {
	if err := w.Writer.Flush(); err != nil {
		w.closer.Close()
		return err
	}
	return w.closer.Close()
}

func newWriter(c *FCGIClient, recType uint8, reqId uint16) *bufWriter {
	s := &streamWriter{c: c, recType: recType, reqId: reqId}
	w := bufio.NewWriterSize(s, maxWrite)
	return &bufWriter{s, w}
}

// streamWriter abstracts out the separation of a stream into discrete records.
// It only writes maxWrite bytes at a time.
type streamWriter struct {
	c       *FCGIClient
	recType uint8
	reqId   uint16
}

func (w *streamWriter) Write(p []byte) (int, error) {
	nn := 0
	for len(p) > 0 {
		n := len(p)
		if n > maxWrite {
			n = maxWrite
		}
		if err := w.c.writeRecord(w.recType, w.reqId, p[:n]); err != nil {
			return nn, err
		}
		nn += n
		p = p[n:]
	}
	return nn, nil
}

func (w *streamWriter) Close() error {
	// send empty record to close the stream
	return w.c.writeRecord(w.recType, w.reqId, nil)
}

func (this *FCGIClient) Close() error {
	select {
	case <- this.broken:
	default: close(this.broken)
	}
	return this.rwc.Close()
}

func (this *FCGIClient) Request(env map[string]string, reqStr string) (retout []byte, reterr []byte, err error) {

	var reqId uint16 = <- this.ctr
	this.ctr <- reqId+1
	
	out := new(bytes.Buffer)
	ber := new(bytes.Buffer)
	ro := respObj{reqId,make(chan struct{}),out,ber}
	this.stream <- ro

	err = this.writeBeginRequest(reqId, uint16(FCGI_RESPONDER), FCGI_KEEP_CONN)
	if err != nil {
		return
	}
	err = this.writePairs(FCGI_PARAMS, reqId, env)
	if err != nil {
		return
	}
	if len(reqStr) > 0 {
		err = this.writeRecord(FCGI_STDIN, reqId, []byte(reqStr))
		if err != nil {
			return
		}
	}
	
	select {
	case <- ro.done:
	case <- this.broken: err=ConnectionBrokenError
	}
	retout = out.Bytes()
	reterr = ber.Bytes()
	
	return
}

/*
 Does the same thing as .Request() but passes all data to rout and rerr Writers, wich is much more efficient.

 The parameter 'rerr' can be nil, as this is checked.
 */
func (this *FCGIClient) RequestIO(env map[string]string, reqStr string, rout, rerr io.Writer) (err error) {

	var reqId uint16 = <- this.ctr
	this.ctr <- reqId+1
	
	ro := respObj{reqId,make(chan struct{}),rout,rerr}
	this.stream <- ro


	err = this.writeBeginRequest(reqId, uint16(FCGI_RESPONDER), FCGI_KEEP_CONN)
	if err != nil {
		return
	}
	err = this.writePairs(FCGI_PARAMS, reqId, env)
	if err != nil {
		return
	}
	if len(reqStr) > 0 {
		err = this.writeRecord(FCGI_STDIN, reqId, []byte(reqStr))
		if err != nil {
			return
		}
	}

	select {
	case <- ro.done:
	case <- this.broken: err=ConnectionBrokenError
	}

	return
}




