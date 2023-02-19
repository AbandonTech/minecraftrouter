import struct
from dataclasses import dataclass
from socket import socket

SEGMENT_BITS = 0x7F
CONTINUE_BIT = 0x80


def read_varint(s: socket) -> tuple[int, bytes]:
    value = 0
    position = 0
    _bytes = bytearray()

    while True:
        byte, = s.recv(1)
        _bytes.append(byte)

        value |= (byte & SEGMENT_BITS) << position

        if (byte & CONTINUE_BIT) == 0:
            break

        position += 7

        if position >= 32:
            raise Exception("Yikes")

    return value, _bytes


@dataclass
class ConnectionFrame:
    protocol_version: int
    server_address: str
    port: int


def read_connection_frame(s: socket) -> tuple[ConnectionFrame, bytes]:
    packet_size, p0 = read_varint(s)
    packet_id, p1 = read_varint(s)

    if packet_id != 0:
        raise RuntimeError("Not a connection frame")

    protocol_ver, p2 = read_varint(s)

    server_addr_size, p3 = read_varint(s)
    server_addr = s.recv(server_addr_size)

    port = s.recv(2)
    next_state, p4 = read_varint(s)
    packet = bytearray([*p0, *p1, *p2, *p3, *server_addr, *port, *p4])

    return ConnectionFrame(protocol_ver, server_addr.decode(), struct.unpack(">H", port)[0]), packet
