package rpc

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
)

type Handler interface {
	GetHandler(path string) (handler HandlerFunc)
	HandleFunc(path string, handlerFunc HandlerFunc)
	ListPath() []string
}

type HandlerFunc func(ResponseWriter, *Request)

type WriteItem struct {
	Data []byte
	Resp chan error
}

type Mux interface {
	Init(io.Reader)
	Handle(*bytes.Buffer) error
	WriteChan() (ch <-chan *WriteItem)
	OnClosed()
}

type Buffer struct {
	*bytes.Buffer
	r io.Reader
}

func NewBuffer(buf *bytes.Buffer) *Buffer {
	return &Buffer{
		Buffer: buf,
		r:      io.MultiReader(buf),
	}
}

func NewBufferString(s string) *Buffer {
	return NewBuffer(bytes.NewBuffer([]byte(s)))
}

func (b *Buffer) Read(buf []byte) (int, error) {
	return b.r.Read(buf)
}

func (b *Buffer) Prepand(r io.Reader) {
	b.r = io.MultiReader(r, b.Buffer)
}

func (b *Buffer) All() []byte {
	data, _ := ioutil.ReadAll(b.r)
	return data
}

type Encoding interface {
	Encode(w BufferWriter, v interface{}) error
	Decode(r *bytes.Reader, v interface{}) error
}

type BufferReader interface {
	io.Reader
	Len() int
	WriteTo(w io.Writer) (int64, error)
	Prepand(r io.Reader)
}

type BufferWriter interface {
	io.Writer
}

type Linker interface {
	Init(net.Conn)
	Handle()
	Protocol() string
}

type ResponseWriter interface {
	Responsef(fmt string, data ...interface{}) error
	Response(data interface{}) error
	Error(err error) error
	Errorf(string, ...interface{}) error
	ErrorInfo(string) error
}

type Protocol interface {
	Read(buf *bytes.Buffer, metaEnc Encoding, p *Packet) error
	Write(metaEnc, bodyEnc Encoding, p *Packet) error
}
