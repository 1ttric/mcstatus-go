package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	i, _ := binary.Varint([]byte{0xff, 0x01})
	fmt.Println(i)
}
