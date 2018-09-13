// Package pcapng provides support for reading packets from PCAP-NG file.
// See https://github.com/pcapng/pcapng for information of the file format.
//
// Made to be compatible with https://github.com/google/gopacket/pcapgo
package pcapng

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
)

func blockName(t uint32) string {
	switch t {
	case blockTypeSHB:
		return "SHB"
	case blockTypeIDB:
		return "IDB"
	case blockTypePB:
		return "PB"
	case blockTypeSPB:
		return "SPB"
	case blockTypeNRB:
		return "NRB"
	case blockTypeISB:
		return "ISB"
	case blockTypeEPB:
		return "EPB"
	case 0x00000BAD, 0x40000BAD:
		return fmt.Sprintf("CB: %x", t)
	}

	return fmt.Sprintf("Unknown: %x", t)
}

const (
	blockTypeSHB uint32 = 0x0A0D0D0A
	blockTypeIDB uint32 = 0x00000001
	blockTypePB  uint32 = 0x00000002
	blockTypeSPB uint32 = 0x00000003
	blockTypeNRB uint32 = 0x00000004
	blockTypeISB uint32 = 0x00000005
	blockTypeEPB uint32 = 0x00000006
	//BlockTypeCB  BlockType = 0x00000BAD or 0x40000BAD
)

const (
	optionEOFOpt  uint16 = 0
	optionTSResol uint16 = 9
)

const magicGzip1 = 0x1f
const magicGzip2 = 0x8b

var magicBigEndian = []byte{0x1A, 0x2B, 0x3C, 0x4D}
var magicLittleEdian = []byte{0x4D, 0x3C, 0x2B, 0x1A}

// ErrPerPacketEncap is the error returned by ReadPacketData when packets have
// different link types (link type of current packet is different from previous
// packets)
var ErrPerPacketEncap = errors.New("doesn't support per-packet encapsulations")

// Reader wraps io.Reader to read packet data in PCAPNG format.
// If the PCAPNG data is gzip compressed it is transparently uncompressed
// by wrapping the given io.Reader with a gzip.Reader.
type Reader struct {
	r         io.Reader
	byteOrder binary.ByteOrder // value of last SHB
	linkType  layers.LinkType  // value of fist IDB

	snapLen uint32 // value of last IDB

	listSHB []*blockSHB // list of SHB

	bh *blockHeader
}

// NewReader returns a new reader object, for reading packet data from
// the given reader. The reader must be open and must contain as
// first block - Section Header Block and Interface Description Block
// before any packet block.
// If the file format is not supported an error is returned
//
//  // Create new reader:
//  f, _ := os.Open("file.pcapng")
//  defer f.Close()
//  r, err := NewReader(f)
//  data, ci, err := r.ReadPacketData()
func NewReader(r io.Reader) (ret *Reader, err error) {
	ret = &Reader{r: r,
		byteOrder: binary.BigEndian,
		listSHB:   make([]*blockSHB, 0),
		bh:        &blockHeader{},
	}

	// from pcapgo
	br := bufio.NewReader(ret.r)
	gzipMagic, err := br.Peek(2)
	if err != nil {
		return nil, err
	}

	if gzipMagic[0] == magicGzip1 && gzipMagic[1] == magicGzip2 {
		if ret.r, err = gzip.NewReader(br); err != nil {
			return nil, err
		}
	} else {
		ret.r = br
	}
	// end from pcapgo

	if err = binary.Read(ret.r, ret.byteOrder, ret.bh); err != nil {
		return nil, err
	}
	if err = ret.processSHB(ret.bh); err != nil {
		return nil, err
	}

	// Read blocks until get first IDB or return error if get EPB/SPB/PB
	for {
		if err = binary.Read(ret.r, ret.byteOrder, ret.bh); err != nil {
			return nil, errors.Wrap(err, "new reader (block header)")
		}

		if ret.bh.Type == blockTypeIDB {
			break
		}

		if ret.bh.Type == blockTypeEPB || ret.bh.Type == blockTypeSPB || ret.bh.Type == blockTypePB {
			return nil, errors.New("need Interface Description Block before first packet block")
		}

		if err = ret.skipBlock(ret.bh); err != nil {
			return nil, errors.Wrap(err, "new reader")
		}
	}

	if err = ret.processIDB(ret.bh); err != nil {
		return nil, err
	}

	idb := ret.lastSHB().getIDB(0)
	ret.linkType = layers.LinkType(idb.linkType)

	return ret, nil
}

// LinkType return link type of first Interface Description block
func (r *Reader) LinkType() layers.LinkType {
	return r.linkType
}

// ReadPacketData read until get packet block (any of EPB, SPB or PB).
// If get SHB or IDB, process it and continue to read.
// Skip any other block types.
// If packets have different link type return error ErrPerPacketEncap.
func (r *Reader) ReadPacketData() (data []byte, ci gopacket.CaptureInfo, err error) {
	for {
		if err = binary.Read(r.r, r.byteOrder, r.bh); err != nil {
			return
		}

		switch r.bh.Type {
		case blockTypeSHB:
			if err = r.processSHB(r.bh); err != nil {
				return
			}
		case blockTypeIDB:
			if err = r.processIDB(r.bh); err != nil {
				return
			}
		case blockTypeEPB, blockTypePB:
			data, ci, err = r.processEPB(r.bh)
			return
		case blockTypeSPB:
			data, ci, err = r.processSPB(r.bh)
			return
		default:
			buf := make([]byte, r.bh.blockRestLength())
			if _, err = io.ReadFull(r.r, buf); err != nil {
				err = errors.Wrapf(err, "block type 0x%x, length %d", r.bh.Type, r.bh.Length)
				return
			}
		}
	}
}

// length of option value padded to 32 bits
func optionLength(l uint16) int {
	return int(l + (4-l%4)%4)
}

type blockHeader struct {
	Type   uint32
	Length uint32
}

func (bh *blockHeader) blockRestLength() int {
	return int(bh.Length - 8) // block header length = 8
}

func (r *Reader) lastSHB() *blockSHB {
	// No need to check length as it is guaranteed when reader was created
	//if len(r.listSHB) == 0 {
	//	return nil
	//}
	return r.listSHB[len(r.listSHB)-1]
}

func (r *Reader) skipBlock(bh *blockHeader) (err error) {
	_, err = io.CopyN(ioutil.Discard, r.r, int64(bh.blockRestLength()))
	err = errors.Wrapf(err, "skip block %s", blockName(bh.Type))
	return
}
