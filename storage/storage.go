package storage

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"time"
)

type Entry struct {
	Id        string
	Prompt    string
	Answer    string
	Embeding  []float64
	CreatedAt time.Time
}

func Float64ToByte(floats []float64) []byte {
	buf := new(bytes.Buffer)
	for _, f := range floats {
		err := binary.Write(buf, binary.LittleEndian, f)
		if err != nil {
			log.Panicf("binary.Write failed: %v", err)
		}
	}
	return buf.Bytes()
}

func (e *Entry) Bytes() ([]byte, error) {
	return json.Marshal(e)
}
