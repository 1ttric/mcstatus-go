package main

import (
	"encoding/binary"
	"fmt"

	"github.com/1ttric/mcstatus-golang/mcstatus"
)

func main() {
	i, _ := binary.Varint([]byte{0xff, 0x01})
	fmt.Println(i)
	c := mcstatus.NewConnection()
	c.test()
}
