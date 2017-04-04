package mcstatus

import (
	"fmt"
	"io"
)

func readVarInt(reader io.Reader) (int, error) {
	numRead := 0
	result := 0
	read := make([]byte, 1, 1)
	for (read[0] & 128) != 0 {
		_, err := reader.Read(read)
		if err != nil {
			return 0, err
		}
		value := read[0] & 127
		result |= (int(value) << uint(7*numRead))

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varint is too large")
		}
	}
	return result, nil
}

func readVarLong(reader io.Reader) (int, error) {
	numRead := 0
	result := 0
	read := make([]byte, 1, 1)
	for (read[0] & 128) != 0 {
		_, err := reader.Read(read)
		if err != nil {
			return 0, err
		}
		value := read[0] & 127
		result |= (int(value) << uint(7*numRead))

		numRead++
		if numRead > 5 {
			return 0, fmt.Errorf("varlong is too large")
		}
	}
	return result, nil
}

func writeVarInt(value int) []byte {
	var output []byte
	for value != 0 {
		temp := byte(value & 127)
		// Note: >>> means that the sign bit is shifted with the rest of the number rather than being left alone
		value >>= 7
		if value != 0 {
			temp |= 128
		}
		output = append(output, byte(temp))
	}
	return output
}

func writeVarLong(value int) []byte {
	return writeVarInt(value)
}
