package minecraft_protocol

import "io"

type Writer struct {
	w io.Writer
}

func (w Writer) Write(packet Packet) {
	//w.w.Write()
}
