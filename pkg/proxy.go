package pkg

import (
	"io"
	"net"
)

func ProxyConnection(c1 net.Conn, c2 net.Conn) error {
	var err error

	go func() {
		_, err = io.Copy(c1, c2)
	}()

	_, err = io.Copy(c2, c1)

	return err
}
