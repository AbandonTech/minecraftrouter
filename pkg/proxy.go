package pkg

import (
	"github.com/rs/zerolog/log"
	"io"
	"net"
)

func ProxyForever(client net.Conn, server net.Conn) {
	// Overwriting name to use this as a contextual log
	log := log.With().
		Stringer("Client", client.RemoteAddr()).
		Stringer("Server", server.RemoteAddr()).
		Logger()

	defer client.Close()
	defer server.Close()

	done := make(chan any)
	var clientWritten, serverWritten int64

	go func() {
		serverWritten, _ = io.Copy(client, server)
		done <- nil
	}()

	go func() {
		clientWritten, _ = io.Copy(server, client)
		done <- nil
	}()

	<-done
	log.Debug().
		Int64("BytesWritten", clientWritten+serverWritten).
		Msg("Finished proxying connection")
}
