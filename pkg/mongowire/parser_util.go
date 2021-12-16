// from https://raw.githubusercontent.com/ma6174/mgosniff/cf913f17c2f681392231629e4c29d90b990a6d2d/utils.go
package mongowire

import (
	"encoding/binary"
	"io"

	jsoniter "github.com/json-iterator/go"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/ioutil"
)

func MustReadUInt8(r io.Reader) (n uint8) {
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return
}

func MustReadUInt32(r io.Reader) (n uint32) {
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return
}

func MustReadInt32(r io.Reader) (n int32) {
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return
}

func ReadInt32(r io.Reader) (n int32, err error) {
	err = binary.Read(r, binary.LittleEndian, &n)
	return
}

func MustReadInt64(r io.Reader) (n int64) {
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		panic(err)
	}
	return
}
func ReadInt64(r io.Reader) *int64 {
	var n int64
	err := binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		panic(err)
	}
	return &n
}

func ReadBytes(r io.Reader, n int) []byte {
	b := make([]byte, n)
	_, err := r.Read(b)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		panic(err)
	}
	return b
}

func ReadCString(r io.Reader) string {
	var b []byte
	var one = make([]byte, 1)
	for {
		_, err := r.Read(one)
		if err != nil {
			panic(err)
		}
		if one[0] == '\x00' {
			break
		}
		b = append(b, one[0])
	}
	return string(b)
}

func ReadOne(r io.Reader) []byte {
	docLen, err := ReadInt32(r)
	if err != nil {
		if err == io.EOF {
			return nil
		}
		panic(err)
	}
	buf := make([]byte, int(docLen))
	binary.LittleEndian.PutUint32(buf, uint32(docLen))
	if _, err := io.ReadFull(r, buf[4:]); err != nil {
		panic(err)
	}
	return buf
}

func ReadDocument(r io.Reader) (m bson.D) {
	if one := ReadOne(r); one != nil {
		err := bson.Unmarshal(one, &m)
		if err != nil {
			panic(err)
		}
	}
	return m
}

func ReadDocuments(r io.Reader) (ms []bson.D) {
	for {
		m := ReadDocument(r)
		if m == nil {
			break
		}
		ms = append(ms, m)
	}
	return
}

func ToJson(v interface{}, l int) string {
	w := ioutil.NewLimitedWriter(make([]byte, l))
	jsoniter.NewEncoder(w).Encode(v)
	return string(w.Get())
}

func hasBit(n int32, pos uint) bool {
	val := n & (1 << pos)
	return (val > 0)
}
