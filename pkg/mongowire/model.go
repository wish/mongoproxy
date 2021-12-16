package mongowire

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/mongo/driver/wiremessage"
)

type WireSerializer interface {
	GetHeader() MessageHeader
	//ToWire() []byte
	WriteTo(io.Writer) error
}

type OP_QUERY_Flags int32

func (f OP_QUERY_Flags) TailableCursor() bool {
	return hasBit(int32(f), 1)
}

func (f OP_QUERY_Flags) SlaveOk() bool {
	return hasBit(int32(f), 2)
}

func (f OP_QUERY_Flags) OplogReplay() bool {
	return hasBit(int32(f), 3)
}

func (f OP_QUERY_Flags) NoCursorTimeout() bool {
	return hasBit(int32(f), 4)
}

func (f OP_QUERY_Flags) AwaitData() bool {
	return hasBit(int32(f), 5)
}

func (f OP_QUERY_Flags) Exhaust() bool {
	return hasBit(int32(f), 6)
}

func (f OP_QUERY_Flags) Partial() bool {
	return hasBit(int32(f), 7)
}

type OP_QUERY struct {
	Header             MessageHeader
	Flags              OP_QUERY_Flags
	FullCollectionName string
	NumberToSkip       int32
	NumberToReturn     int32

	Query                bson.D
	ReturnFieldsSelector bson.D
}

func (m *OP_QUERY) GetHeader() MessageHeader {
	return m.Header
}

func (m *OP_QUERY) FromWire(r io.Reader) {
	m.Flags = OP_QUERY_Flags(MustReadInt32(r))
	m.FullCollectionName = ReadCString(r)
	m.NumberToSkip = MustReadInt32(r)
	m.NumberToReturn = MustReadInt32(r)
	m.Query = ReadDocument(r)
	m.ReturnFieldsSelector = ReadDocument(r)
}

type OP_KILL_CURSORS struct {
	Header            MessageHeader
	ZERO              int32
	NumberOfCursorIDs int32
	CursorIDs         []int64
}

func (m *OP_KILL_CURSORS) GetHeader() MessageHeader {
	return m.Header
}

func (m *OP_KILL_CURSORS) FromWire(r io.Reader) {
	m.ZERO = MustReadInt32(r)
	m.NumberOfCursorIDs = MustReadInt32(r)
	m.CursorIDs = make([]int64, m.NumberOfCursorIDs)
	for i := int32(0); i < m.NumberOfCursorIDs; i++ {
		m.CursorIDs[i] = MustReadInt64(r)
	}
}

type OP_GETMORE struct {
	Header             MessageHeader
	Flags              int32
	FullCollectionName string
	NumberToReturn     int32
	CursorID           int64
}

func (m *OP_GETMORE) FromWire(r io.Reader) {
	m.Flags = MustReadInt32(r)
	m.FullCollectionName = ReadCString(r)
	m.NumberToReturn = MustReadInt32(r)
	m.CursorID = MustReadInt64(r)
}

func (m *OP_GETMORE) GetHeader() MessageHeader {
	return m.Header
}

type OP_REPLY struct {
	Header         MessageHeader
	Flags          int32
	CursorID       int64
	StartingFrom   int32
	NumberReturned int32
	Documents      []bson.D // interface?
}

func (m *OP_REPLY) GetHeader() MessageHeader {
	return m.Header
}

func (o *OP_REPLY) ToWire() ([]byte, error) {
	bodyBuf := bytes.NewBuffer(nil)
	binary.Write(bodyBuf, binary.LittleEndian, &o.Flags)
	binary.Write(bodyBuf, binary.LittleEndian, &o.CursorID)
	binary.Write(bodyBuf, binary.LittleEndian, &o.StartingFrom)
	binary.Write(bodyBuf, binary.LittleEndian, &o.NumberReturned)
	for _, doc := range o.Documents {
		b, err := bson.Marshal(doc)
		if err != nil {
			return nil, err
		}
		if _, err := bodyBuf.Write(b); err != nil {
			return nil, err
		}
	}

	o.Header.MessageLength = int32(bodyBuf.Len()) + HeaderLen

	hb, err := o.Header.ToWire()
	if err != nil {
		return nil, err
	}

	return append(hb, bodyBuf.Bytes()...), nil
}

func (o *OP_REPLY) WriteTo(w io.Writer) error {
	b, err := o.ToWire()
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWrite
	}
	return nil
}

type OP_MSG_Flags int32

func (f OP_MSG_Flags) ChecksumPresent() bool {
	return hasBit(int32(f), 0)
}

func (f OP_MSG_Flags) MoreToCome() bool {
	return hasBit(int32(f), 1)
}

type OP_MSG struct {
	Header   MessageHeader
	Flags    OP_MSG_Flags
	Sections []MSGSection
	Checksum uint32
}

func (m *OP_MSG) GetHeader() MessageHeader {
	return m.Header
}

func (o *OP_MSG) FromWire(r io.Reader, crc *Crc32c, msgLen int) error {
	o.Flags = OP_MSG_Flags(MustReadUInt32(r))

	var checksumLength int
	if o.Flags.ChecksumPresent() {
		checksumLength = 4
	}

	cR := &CountingReader{reader: r, BytesRead: 4}
	for cR.BytesRead < msgLen-checksumLength {
		t := ReadBytes(cR, 1)
		if t == nil {
			break
		}
		switch t[0] {
		case 0: // body
			o.Sections = append(o.Sections, MSGSection_Body{ReadDocument(cR)})
		case 1:
			sectionSize := MustReadInt32(cR)
			sectionSize -= 4 // the sectionSize counts towards the length, so lets remove
			r1 := io.LimitReader(cR, int64(sectionSize))
			o.Sections = append(o.Sections, MSGSection_DocumentSequence{
				Size:               sectionSize,
				SequenceIdentifier: ReadCString(r1),
				Documents:          ReadDocuments(r1),
			})
		default:
			msg := fmt.Sprintf("unknown body kind=%v", t[0])
			panic(msg)
		}
	}

	if o.Flags.ChecksumPresent() {
		if crc == nil {
			err := fmt.Errorf("CRC checksum present but crc not computed")
			logrus.Error(err)
			return err
		}
		crcGen := crc.GetCrc()
		o.Checksum = MustReadUInt32(r)
		if crcGen != o.Checksum {
			err := fmt.Errorf("crc Check failed, Generated=%v:(0x%x), Got=%v:(0x%x)", crcGen, crcGen, o.Checksum, o.Checksum)
			logrus.Error(err)
			return err
		}
		logrus.Debugf("CRC:  Generated=%v:(0x%x), Got=%v:(0x%x)\n", crcGen, crcGen, o.Checksum, o.Checksum)
	}
	return nil
}

func (o *OP_MSG) ToWire() ([]byte, error) {
	bodyBuf := bytes.NewBuffer(nil)
	binary.Write(bodyBuf, binary.LittleEndian, &o.Flags)
	for _, section := range o.Sections {
		switch sectionTyped := section.(type) {
		case MSGSection_Body:
			if _, err := bodyBuf.Write([]byte{0}); err != nil {
				return nil, err
			}
			b, err := bson.Marshal(sectionTyped.Document)
			if err != nil {
				return nil, err
			}
			if _, err := bodyBuf.Write(b); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unknown section type")
		}
	}

	o.Header.MessageLength = int32(bodyBuf.Len()) + HeaderLen

	hb, err := o.Header.ToWire()
	if err != nil {
		return nil, err
	}

	return append(hb, bodyBuf.Bytes()...), nil
}

func (o *OP_MSG) WriteTo(w io.Writer) error {
	b, err := o.ToWire()
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWrite
	}
	return nil
}

type MSGSection interface {
	MSGSection()
}

type MSGSection_Body struct {
	Document bson.D
}

func (MSGSection_Body) MSGSection() {}

type MSGSection_DocumentSequence struct {
	Size               int32
	SequenceIdentifier string
	Documents          []bson.D
}

func (MSGSection_DocumentSequence) MSGSection() {}

type OP_COMPRESSED struct {
	Header            MessageHeader
	OriginalOpcode    OpCode
	UncompressedSize  int32
	CompressorID      wiremessage.CompressorID
	CompressedMessage []byte
}

func (m *OP_COMPRESSED) GetHeader() MessageHeader {
	return m.Header
}

func (o *OP_COMPRESSED) FromWire(r io.Reader) error {
	o.OriginalOpcode = OpCode(MustReadInt32(r))
	o.UncompressedSize = MustReadInt32(r)
	o.CompressorID = wiremessage.CompressorID(ReadBytes(r, 1)[0])
	o.CompressedMessage = ReadBytes(r, int(o.Header.MessageLength-25)) // header (16) + original opcode (4) + uncompressed size (4) + compressor ID (1)

	return nil
}

func (o *OP_COMPRESSED) ToWire() ([]byte, error) {
	bodyBuf := bytes.NewBuffer(nil)
	binary.Write(bodyBuf, binary.LittleEndian, &o.OriginalOpcode)
	binary.Write(bodyBuf, binary.LittleEndian, &o.UncompressedSize)
	binary.Write(bodyBuf, binary.LittleEndian, &o.CompressorID)
	if _, err := bodyBuf.Write(o.CompressedMessage); err != nil {
		return nil, err
	}

	o.Header.MessageLength = int32(bodyBuf.Len()) + HeaderLen

	hb, err := o.Header.ToWire()
	if err != nil {
		return nil, err
	}

	return append(hb, bodyBuf.Bytes()...), nil
}

func (o *OP_COMPRESSED) WriteTo(w io.Writer) error {
	b, err := o.ToWire()
	if err != nil {
		return err
	}
	n, err := w.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return errWrite
	}
	return nil
}
