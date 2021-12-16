package mongowire

import (
	"errors"
	"fmt"
	"hash/crc32"
)

type Crc32c struct {
	table *crc32.Table
	crc   uint32
}

func (c *Crc32c) Init() {
	c.table = crc32.MakeTable(crc32.Castagnoli)
}

func (c *Crc32c) UpdateCrc(msg []byte) {
	c.crc = crc32.Update(c.crc, c.table, msg)
}

func (c *Crc32c) GetCrc() uint32 {
	return c.crc
}

func (c *Crc32c) CheckCrc(crc uint32) bool {
	return c.crc == crc
}

func (c *Crc32c) String() string {
	return fmt.Sprintf("CRC: %v:(0x%x)", c.crc, c.crc)
}

func (c *Crc32c) Write(msg []byte) (int, error) {
	if c == nil {
		return 0, errors.New("nil crc interface")
	}
	c.UpdateCrc(msg)
	return len(msg), nil
}
