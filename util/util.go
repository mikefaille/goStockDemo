package util

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"unsafe"
)

//https://github.com/golang/go/issues/2632#issuecomment-66061057
func Float64frombytes3(data []byte) (theFloat float64, err error) {
	//	data := make([]byte, len(byt))
	//	copy(data, byt)

	unsafeString := func(b []byte) string {
		return *(*string)(unsafe.Pointer(&b))
	}

	theFloat, err = strconv.ParseFloat(unsafeString(data), 64)

	// theFloat, err = strconv.ParseFloat(string(data), 64)

	Check(err)

	//	fmt.Println("theFloat", theFloat)
	// if theFloat == 0. {

	// 	err = errors.New("out = 0")

	// }

	return
}

//https://github.com/golang/go/issues/2632#issuecomment-66061057
func Float64frombytes4(byt []byte) (theFloat float64, err error) {
	//	data := make([]byte, len(byt))
	//	copy(data, byt)
	if len(byt) < 8 {
		data := make([]byte, 8)
		copy(data, byt)
	}
	bits := binary.LittleEndian.Uint64(byt)
	theFloat = math.Float64frombits(bits)
	return

}

func Check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
