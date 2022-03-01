package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/Qitmeer/meerevm/common"
	address2 "github.com/Qitmeer/qng-core/core/address"
)

func LittleHexToUint64(hexStr string) (uint64, error) {
	if len(hexStr) == 1 {
		hexStr = "0" + hexStr
	}
	src, err := hex.DecodeString(hexStr)
	if err != nil {
		return 0, err
	}
	dst := make([]byte, 8)
	copy(dst, src)
	var number uint64
	bytesBuffer := bytes.NewBuffer(dst)
	err = binary.Read(bytesBuffer, binary.LittleEndian, &number)
	return number, err
}

func IsPkAddress(address string) bool {
	return len(address) == 53
}

func PkAddressToAddress(addr string) (string, error) {
	address, err := address2.DecodeAddress(addr)
	if err != nil {
		return "", err
	}
	return address.Encode(), nil
}

func PkAddressToEVMAddress(addr string) (string, error) {
	address, err := address2.DecodeAddress(addr)
	if err != nil {
		return "", err
	}
	pkHex := hex.EncodeToString(address.Script())
	evmAddr, err := common.NewMeerEVMAddress(pkHex)
	if err != nil {
		return "", err
	}
	return evmAddr.String(), nil
}
