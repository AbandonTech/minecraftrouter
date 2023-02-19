import select
import socket
import socketserver

import httpx

from router.config import API_URL
from router.minecraft_protocol import read_connection_frame

BUFFER_SIZE = 1024


def create_proxy_protocol_header(src_addr: tuple[str, int], dst_addr: tuple[str, int]) -> bytes:
    return f"PROXY TCP4 {src_addr[0]} {dst_addr[0]} {src_addr[1]} {dst_addr[1]}\r\n".encode()


def resolve_minecraft_server(servername: str, port: int) -> tuple[str, int]:
    server_mappings = httpx.get(f"{API_URL}/service/mapping").json()
    print("server_mappings", server_mappings)

    resolved_mapping = server_mappings.get(f"{servername}:{port}") or server_mappings.get(servername)

    if resolved_mapping is None:
        raise Exception(f"Couldn't find a mapping for {servername}:{port}")

    return resolved_mapping["address"], int(resolved_mapping["port"])


class ProxyConnection(socketserver.StreamRequestHandler):
    def setup(self):
        super().setup()

        connection_frame, raw_connection_frame = read_connection_frame(self.connection)
        minecraft_server_addr = resolve_minecraft_server(connection_frame.server_address, connection_frame.port)

        self.remote_connection = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.remote_connection.connect(minecraft_server_addr)

        header = create_proxy_protocol_header(self.connection.getpeername(), minecraft_server_addr)
        self.remote_connection.send(header)
        self.remote_connection.sendall(raw_connection_frame)

    def handle(self):

        try:
            self.handle_forever()
        except (BrokenPipeError, ConnectionResetError):
            ...
        finally:
            self.finish()

    def handle_forever(self):
        while True:
            read_ready, _, _ = select.select([self.remote_connection, self.connection], [], [], self.timeout)

            for connection in read_ready:
                data = connection.recv(BUFFER_SIZE)

                if not data:
                    return

                if connection is self.connection:
                    self.remote_connection.send(data)
                else:
                    self.connection.send(data)

    def finish(self):
        super().finish()
        self.remote_connection.close()


class Server(socketserver.ThreadingTCPServer):
    timeout = 1
    daemon_threads = True


print("hosting on", ("0.0.0.0", 25565))
with Server(server_address=("0.0.0.0", 25565), RequestHandlerClass=ProxyConnection) as server:
    try:
        server.serve_forever()
    finally:
        server.shutdown()
