package api

import (
	"bytes"
	"encoding/gob"
	"strconv"
)

func Encode(p *Product) []byte {
	var buff bytes.Buffer
	// Кодирование значения
	enc := gob.NewEncoder(&buff)
	enc.Encode(p)
	// fmt.Printf("%X\n", buff.Bytes())
	return buff.Bytes()
}

func Decode(b []byte) *Product {
	buff := bytes.NewBuffer(b)
	P := Product{}
	// Декодирование значения
	dec := gob.NewDecoder(buff)
	dec.Decode(&P)
	// fmt.Println(out.String())
	return &P
}

func IntToByte(id int) []byte {
	return []byte(strconv.FormatInt(int64(id), 10))
}
