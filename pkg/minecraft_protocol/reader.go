package minecraft_protocol

import "io"

type Reader struct {
	io.Reader
}

func (r Reader) Read() Packet {
	return []byte{}
}
