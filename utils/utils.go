package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
)


func LittleHexToUint64(hexStr string)(uint64, error){
	src, err := hex.DecodeString(hexStr)
	if err != nil{
		return 0, err
	}
	dst := make([]byte, 8)
	copy(dst, src)
	var number uint64
	bytesBuffer := bytes.NewBuffer(dst)
	err = binary.Read(bytesBuffer, binary.LittleEndian, &number)
	return number, err
}