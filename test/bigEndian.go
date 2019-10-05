package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func main() {
	var i uint16 = 1

	// ??
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	fmt.Printf("LittleEndian(%d) :", i)
	for _, bin := range b {
		fmt.Printf("%02X ", bin)
	}
	fmt.Printf("\n")

	//??
	fmt.Printf("BigEndian(%d) :", i)
	binary.BigEndian.PutUint16(b, i)
	for _, bin := range b {
		fmt.Printf("%02X ", bin)
	}
	fmt.Printf("\n")

	//[]byte 2 uint32
	bytesBuffer := bytes.NewBuffer(b)
	var j,k uint16
	binary.Read(bytesBuffer, binary.LittleEndian, &j)
	fmt.Println("j = ", j)
	binary.Read(bytesBuffer, binary.BigEndian, &k)
	fmt.Println("k = ", k)

	//bytesBuffer1 := bytes.NewBuffer(b)




}