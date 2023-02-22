package pkg

import (
	"io"
	"net"
)

func ProxyConnection(c1 net.Conn, c2 net.Conn) {
	go io.Copy(c1, c2)
	io.Copy(c2, c1)
}
