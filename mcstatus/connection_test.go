package mcstatus

import (
	"reflect"
	"strings"
	"testing"
)

func TestFlush(t *testing.T) {
	c := NewConnection()
	c.sent = []byte{0x7F, 0xAA, 0xBB}
	data := c.Flush()
	if !reflect.DeepEqual(data, []byte{0x7F, 0xAA, 0xBB}) {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA, 0xBB}, data)
	}
}

func TestReceive(t *testing.T) {
	c := NewConnection()
	c.Receive([]byte{0x7F})
	c.Receive([]byte{0xAA, 0xBB})
	if !reflect.DeepEqual(c.received, []byte{0x7F, 0xAA, 0xBB}) {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA, 0xBB}, c.received)
	}
}

func TestRemaining(t *testing.T) {
	c := NewConnection()
	c.Receive([]byte{0x7F})
	c.Receive([]byte{0xAA, 0xBB})
	rem := c.Remaining()
	if rem != 3 {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA, 0xBB}, rem)
	}
}

func TestSend(t *testing.T) {
	c := NewConnection()
	c.Write([]byte{0x7F})
	c.Write([]byte{0xAA, 0xBB})
	data := c.Flush()
	if !reflect.DeepEqual(data, []byte{0x7F, 0xAA, 0xBB}) {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA, 0xBB}, data)
	}
}

func TestRead(t *testing.T) {
	c := NewConnection()
	c.Receive([]byte{0x7F, 0xAA, 0xBB})
	data, err := c.Read(2)
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if !reflect.DeepEqual(data, []byte{0x7F, 0xAA}) {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA}, data)
	}
	data, err = c.Read(1)
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if !reflect.DeepEqual(data, []byte{0xBB}) {
		t.Errorf("Expected %q, got %q", []byte{0xBB}, data)
	}
}

func TestReadSimpleVarInt(t *testing.T) {
	expected := 15

	c := NewConnection()
	c.Receive([]byte{0x0F})
	i, err := c.ReadVarInt()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteSimpleVarInt(t *testing.T) {
	expected := []byte{0x0F}

	c := NewConnection()
	err := c.WriteVarInt(15)
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}

	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadBigVarInt(t *testing.T) {
	expected := 34359738367

	c := NewConnection()
	c.Receive([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x7F})
	i, err := c.ReadVarInt()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteBigVarInt(t *testing.T) {
	expected := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0x07}

	c := NewConnection()
	err := c.WriteVarInt(2147483647)
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}

	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadInvalidVarInt(t *testing.T) {
	expected := "server sent a varint that was too big"

	c := NewConnection()
	c.Receive([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x80})
	_, err := c.ReadVarInt()
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expected)
	} else if strings.Compare(err.Error(), expected) != 0 {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestWriteInvalidVarInt(t *testing.T) {
	expected := "value is too big to send in a varint"

	c := NewConnection()
	err := c.WriteVarInt(34359738368)
	if err == nil {
		t.Errorf("Expected error '%s', got nil", expected)
	} else if strings.Compare(err.Error(), expected) != 0 {
		t.Errorf("Expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestReadUTF(t *testing.T) {
	expected := "Hello, world!"

	c := NewConnection()
	c.Receive([]byte{0x0D, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21})
	str, err := c.ReadUTF()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if strings.Compare(str, expected) != 0 {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestWriteUTF(t *testing.T) {
	expected := []byte{0x0D, 0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21}

	c := NewConnection()
	c.WriteUTF("Hello, world!")
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadEmptyUTF(t *testing.T) {
	expected := []byte{0x00}

	c := NewConnection()
	c.WriteUTF("")
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadASCII(t *testing.T) {
	expected := "Hello, world!"

	c := NewConnection()
	c.Receive([]byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21, 0x00})
	str, err := c.ReadASCII()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if strings.Compare(str, expected) != 0 {
		t.Errorf("Expected '%s', got '%s'", expected, str)
	}
}

func TestWriteASCII(t *testing.T) {
	expected := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x2C, 0x20, 0x77, 0x6F, 0x72, 0x6C, 0x64, 0x21, 0x00}

	c := NewConnection()
	c.WriteASCII("Hello, world!")
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadEmptyASCII(t *testing.T) {
	expected := []byte{0x00}

	c := NewConnection()
	c.WriteASCII("")
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadShortNegative(t *testing.T) {
	expected := int16(-32768)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00})
	i, err := c.ReadShort()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteShortNegative(t *testing.T) {
	expected := []byte{0x80, 0x00}

	c := NewConnection()
	c.WriteShort(int16(-32768))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadShortPositive(t *testing.T) {
	expected := int16(32767)

	c := NewConnection()
	c.Receive([]byte{0x7F, 0xFF})
	i, err := c.ReadShort()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteShortPositive(t *testing.T) {
	expected := []byte{0x7F, 0xFF}

	c := NewConnection()
	c.WriteShort(int16(32767))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadUShortPositive(t *testing.T) {
	expected := uint16(32768)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00})
	i, err := c.ReadUshort()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteUShortPositive(t *testing.T) {
	expected := []byte{0x80, 0x00}

	c := NewConnection()
	c.WriteUshort(uint16(32768))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadIntNegative(t *testing.T) {
	expected := int32(-2147483648)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00, 0x00, 0x00})
	i, err := c.ReadInt()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteIntNegative(t *testing.T) {
	expected := []byte{0x80, 0x00, 0x00, 0x00}

	c := NewConnection()
	c.WriteInt(int32(-2147483648))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadIntPositive(t *testing.T) {
	expected := int32(2147483647)

	c := NewConnection()
	c.Receive([]byte{0x7F, 0xFF, 0xFF, 0xFF})
	i, err := c.ReadInt()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteIntPositive(t *testing.T) {
	expected := []byte{0x7F, 0xFF, 0xFF, 0xFF}

	c := NewConnection()
	c.WriteInt(int32(2147483647))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadUIntPositive(t *testing.T) {
	expected := uint32(2147483648)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00, 0x00, 0x00})
	i, err := c.ReadUint()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteUIntPositive(t *testing.T) {
	expected := []byte{0x80, 0x00, 0x00, 0x00}

	c := NewConnection()
	c.WriteUint(uint32(2147483648))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadLongNegative(t *testing.T) {
	expected := int64(-9223372036854775808)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	i, err := c.ReadLong()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteLongNegative(t *testing.T) {
	expected := []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	c := NewConnection()
	c.WriteLong(int64(-9223372036854775808))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadLongPositive(t *testing.T) {
	expected := int64(9223372036854775807)

	c := NewConnection()
	c.Receive([]byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})
	i, err := c.ReadLong()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteLongPositive(t *testing.T) {
	expected := []byte{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	c := NewConnection()
	c.WriteLong(int64(9223372036854775807))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadULongPositive(t *testing.T) {
	expected := uint64(9223372036854775808)

	c := NewConnection()
	c.Receive([]byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	i, err := c.ReadULong()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if i != expected {
		t.Errorf("Expected %d, got %d", expected, i)
	}
}

func TestWriteULongPositive(t *testing.T) {
	expected := []byte{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

	c := NewConnection()
	c.WriteULong(uint64(9223372036854775808))
	data := c.Flush()
	if !reflect.DeepEqual(data, expected) {
		t.Errorf("Expected %q, got %q", expected, data)
	}
}

func TestReadBuffer(t *testing.T) {
	c := NewConnection()

	c.Receive([]byte{0x02, 0x7F, 0xAA})
	buffer, err := c.ReadBuffer()
	if err != nil {
		t.Errorf("Encountered error: %s", err.Error())
	}
	if !reflect.DeepEqual(buffer.received, []byte{0x7F, 0xAA}) {
		t.Errorf("Expected %q, got %q", []byte{0x7F, 0xAA}, buffer.received)
	}
	data := c.Flush()
	if !reflect.DeepEqual(data, []byte{}) {
		t.Errorf("Expected %q, got %q", []byte{}, data)
	}
}

func TestWriteBuffer(t *testing.T) {
	c := NewConnection()

	buffer := NewConnection()
	buffer.Write([]byte{0x7F, 0xAA})
	c.WriteBuffer(buffer)
	data := c.Flush()
	if !reflect.DeepEqual(data, []byte{0x02, 0x7F, 0xAA}) {
		t.Errorf("Expected %q, got %q", []byte{0x02, 0x7F, 0xAA}, data)
	}

}
