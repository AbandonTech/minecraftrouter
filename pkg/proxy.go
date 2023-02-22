package pkg

import (
	"io"
	"net"
)

func ProxyForever(c1 net.Conn, c2 net.Conn) {
	defer c1.Close()
	defer c2.Close()

	done := make(chan any)

	go func() {
		_, _ = io.Copy(c1, c2)
		done <- nil
	}()

	go func() {
		_, _ = io.Copy(c2, c1)
		done <- nil
	}()

	<-done
}
