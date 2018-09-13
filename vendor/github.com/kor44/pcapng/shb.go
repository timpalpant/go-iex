package pcapng

import (
	"encoding/binary"
	"io"
	"io/ioutil"

	"github.com/pkg/errors"
)

type blockSHB struct {
	byteOrder binary.ByteOrder
	listIDB   []*blockIDB
}

func (shb *blockSHB) getIDB(i int) *blockIDB {
	if len(shb.listIDB) > i {
		return shb.listIDB[i]
	}
	return nil
}

func (r *Reader) processSHB(bh *blockHeader) (err error) {
	// Minimum SHB size: ByteOrderMagic(4 byte) + Version (4 bytes) + SectionLen(8 bytes) + Block Total Length (4 bytes) = 20

	magic := make([]byte, 4)
	if _, err = io.ReadFull(r.r, magic); err != nil {
		return errors.Wrap(err, "read SHB")
	}

	shb := blockSHB{
		listIDB: make([]*blockIDB, 0),
	}
	r.listSHB = append(r.listSHB, &shb)

	if magic[0] == magicLittleEdian[0] && magic[1] == magicLittleEdian[1] &&
		magic[2] == magicLittleEdian[2] && magic[3] == magicLittleEdian[3] {
		shb.byteOrder = binary.LittleEndian
	} else if magic[0] == magicBigEndian[0] && magic[1] == magicBigEndian[1] &&
		magic[2] == magicBigEndian[2] && magic[3] == magicBigEndian[3] {
		shb.byteOrder = binary.BigEndian
	} else {
		return errors.Errorf("read SHB, unknown magic %x", magic)
	}

	if r.byteOrder != shb.byteOrder {
		r.byteOrder = shb.byteOrder
		b := make([]byte, 4)
		b[0] = byte(bh.Length)
		b[1] = byte(bh.Length >> 8)
		b[2] = byte(bh.Length >> 16)
		b[3] = byte(bh.Length >> 24)
		bh.Length = uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
	}

	//skip to end of section
	if _, err = io.CopyN(ioutil.Discard, r.r, int64(bh.Length)-12); err != nil {
		return errors.Wrap(err, "read SHB, not enough data")
	}

	return
}
