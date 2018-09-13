package pcapng

import (
	"io"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/pkg/errors"
)

func (r *Reader) processEPB(bh *blockHeader) (data []byte, ci gopacket.CaptureInfo, err error) {
	if bh.blockRestLength() < 24 {
		// Minimum EPB size: IntefaceID(4 byte) + Timestamp (8 bytes) + CaptureLen(4 bytes) + OrigLength(4 bytes) +
		//                   Block Total Length (4 bytes) = 24
		err = errors.Errorf("read %s, incorrect length %d", blockName(bh.Type), bh.blockRestLength())
		return
	}

	buf := make([]byte, bh.blockRestLength())
	if _, err = io.ReadFull(r.r, buf); err != nil {
		err = errors.Wrap(err, "read EPB")
		return
	}

	if bh.Type == blockTypeEPB {
		ci.InterfaceIndex = int(r.byteOrder.Uint32(buf[:4]))
	} else {
		ci.InterfaceIndex = int(r.byteOrder.Uint32(buf[:2]))
	}

	idb, err := r.checkIDBLinkType(ci.InterfaceIndex)
	if err != nil {
		//err = errors.Wrap(err, "read EPB")
		return
	}

	tsHigh := r.byteOrder.Uint32(buf[4:8])
	tsLow := r.byteOrder.Uint32(buf[8:12])
	ci.Timestamp = timestamp(tsHigh, tsLow, idb.timeUnitsPerSecond)

	ci.CaptureLength = int(r.byteOrder.Uint32(buf[12:16]))
	ci.Length = int(r.byteOrder.Uint32(buf[16:20]))

	if len(buf) > 4 {
		buf = buf[:len(buf)-4]
	} else {
		err = errors.New("read SPB: not enough data length")
		return
	}

	data = buf[20 : 20+ci.CaptureLength]
	return
}

// Simple Packet Block
func (r *Reader) processSPB(bh *blockHeader) (data []byte, ci gopacket.CaptureInfo, err error) {
	if bh.blockRestLength() < 8 {
		// Minimum SPB size:  OrigLength(4 bytes) +  Block Total Length (4 bytes) = 8
		err = errors.Errorf("read SPB, incorrect length %d", bh.blockRestLength())
		return
	}

	buf := make([]byte, bh.blockRestLength())
	if _, err = io.ReadFull(r.r, buf); err != nil {
		err = errors.Wrap(err, "read SPB")
		return
	}

	// Simple Packet Block does not contains timestamp info
	ci.Timestamp = time.Unix(0, 0)

	// original length
	origLen := r.byteOrder.Uint32(buf[:4])
	buf = buf[4:]

	// Simple Packet Blocks have been captured on the interface previously
	// specified in the first Interface Description Block
	idb, err := r.checkIDBLinkType(0)
	if err != nil {
		//err = errors.Wrap(err, "read SPB")
		return
	}

	if len(buf) > 4 {
		buf = buf[:len(buf)-4]
	} else {
		err = errors.New("read SPB: not enough data length")
		return
	}
	// If snapLen less than original length, snapLen MUST be used
	// to determine the size of the Packet Data field length.
	ci.CaptureLength = int(origLen)
	if int(origLen) <= len(buf) {
		data = buf[:origLen]
	} else if idb.snapLen != 0 && int(idb.snapLen) < len(buf) {
		data = buf[:idb.snapLen]
	} else {
		data = buf
	}
	ci.Length = len(data)

	return
}

func (r *Reader) checkIDBLinkType(idx int) (*blockIDB, error) {
	idb := r.lastSHB().getIDB(idx)
	if idb == nil {
		return nil, errors.New("lost IDB")
	}

	//&& idb.linkType != 0
	if r.linkType != layers.LinkTypeNull && r.linkType != layers.LinkType(idb.linkType) {
		return idb, ErrPerPacketEncap
	} else if r.linkType == layers.LinkTypeNull {
		r.linkType = layers.LinkType(idb.linkType)
	}

	return idb, nil
}

func timestamp(tsHigh uint32, tsLow uint32, timeUnitsPerSecond uint64) time.Time {
	ts := ((uint64(tsHigh)) << 32) | uint64(tsLow)
	secs := int32(ts / timeUnitsPerSecond)
	nsecs := int32(((ts % timeUnitsPerSecond) * 1000000000) / timeUnitsPerSecond)
	return time.Unix(int64(secs), int64(nsecs))
}
