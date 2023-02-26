package reader

import (
	"errors"
	"io"
)

const defaultBufSize = 4096 * 4
const maxBytesVarInt = 5

type MinecraftPacketReader struct {
	buf      []byte
	reader   io.Reader
	readHead int
}

func (m *MinecraftPacketReader) readVarInt() (int, []byte) {
	result := 0
	buf := make([]byte, 5)

	for i := 0; i < maxBytesVarInt; i++ {
		_, err := m.reader.Read(buf[i : i+1])
		if err != nil {
			return 0, nil
		}

		value := buf[i] & 0x7F
		result |= int(value) << (7 * i)
		if (buf[i] & 0x80) == 0 {
			return result, buf[:i+1]
		}
	}

	return -1, nil
}

func (m *MinecraftPacketReader) ReadPacket() ([]byte, error) {
	length, data := m.readVarInt()
	println("varint bytes", len(data))
	if data == nil {
		return nil, errors.New("unable to read varint")
	}

	packet := make([]byte, length)
	r, err := m.reader.Read(packet)
	if err != nil {
		return nil, err
	}

	if r != length {
		return nil, errors.New("cannot read entire packet")
	}

	return append(data, packet...), nil
}

func NewMinecraftReader(r io.Reader) MinecraftPacketReader {
	return MinecraftPacketReader{
		reader: r,
	}
}
