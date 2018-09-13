package pcapng

import (
	"io"

	"github.com/pkg/errors"
)

type blockIDB struct {
	linkType           uint16
	snapLen            uint32
	timeUnitsPerSecond uint64
}

func (r *Reader) processIDB(bh *blockHeader) (err error) {
	if bh.blockRestLength() < 12 {
		// Minimum IDB size: LinkType(2 byte) + Reserved (2 bytes) + SnapLen(4 bytes) + Block Total Length (4 bytes) = 12
		return errors.Errorf("read IDB, incorrect length %d", bh.blockRestLength())
	}
	buf := make([]byte, bh.blockRestLength())

	if _, err = io.ReadFull(r.r, buf); err != nil {
		return errors.Wrap(err, "read IDB")
	}

	var idb blockIDB
	idb.timeUnitsPerSecond = 1000000 /* default = 10^6 */

	idb.linkType = r.byteOrder.Uint16(buf[:2])

	shb := r.lastSHB()
	if shb == nil {
		return errors.New("need Section Header Block")
	}
	shb.listIDB = append(shb.listIDB, &idb)

	idb.snapLen = r.byteOrder.Uint32(buf[4:8])

	buf = buf[8:]

	for len(buf) > 4 || err != nil {
		code := r.byteOrder.Uint16(buf[:2])
		l := r.byteOrder.Uint16(buf[2:4])
		buf = buf[4:]

		var optData []byte
		lpad := optionLength(l)
		if len(buf) >= lpad {
			if int(l) < lpad {
				optData = buf[:l]
			} else {
				optData = buf[:lpad]
			}
			buf = buf[lpad:]
		} else {
			return errors.Errorf("IDB, option data length not enough %d", len(buf))
		}

		if code == optionEOFOpt {
			break
		}
		if code == optionTSResol {
			if l != 1 {
				return errors.Errorf("IDB, ts_resol length %d not 1 as expected", l)
			}
			tsResol := optData[0]
			var base uint64
			if (tsResol & 0x80) > 0 {
				base = 2
			} else {
				base = 10
			}
			exponent := tsResol & 0x7f

			if ((base == 2) && (exponent < 64)) || ((base == 10) && (exponent < 20)) {
				var result uint64 = 1
				var i byte
				for i = 0; i < exponent; i++ {
					result *= base
				}
				idb.timeUnitsPerSecond = result
			} else {
				idb.timeUnitsPerSecond = 0xFFFFFFFFFFFFFFFF
			}
			break
		}
	}
	return
}
