package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

type tiffReader struct {
	data  []byte
	order binary.ByteOrder
}

type tiffEntry struct {
	tag   uint16
	typ   uint16
	count uint32
	raw   []byte
}

func readEXIFMetadata(path string) map[string]string {
	out := map[string]string{}
	ext := strings.ToLower(filepath.Ext(path))

	var tiff []byte
	switch ext {
	case ".jpg", ".jpeg", ".jpe", ".jfif":
		tiff = extractJPEGExif(path)
	case ".tif", ".tiff":
		data, err := os.ReadFile(path)
		if err == nil {
			tiff = data
		}
	default:
		return out
	}
	if len(tiff) < 8 {
		return out
	}

	tr, ok := newTIFFReader(tiff)
	if !ok {
		return out
	}

	ifd0 := tr.u32(4)
	entries := tr.readIFD(ifd0)
	var exifOffset, gpsOffset uint32

	for _, e := range entries {
		switch e.tag {
		case 0x010e:
			setMeta(out, "title", entryString(e))
		case 0x013b:
			setMeta(out, "author", entryString(e))
		case 0x8298:
			setMeta(out, "copyright", entryString(e))
		case 0x9c9b:
			setMeta(out, "title", decodeXPString(e.raw))
		case 0x9c9f:
			setMeta(out, "subject", decodeXPString(e.raw))
		case 0x011a:
			if v, ok := tr.entryRational(e, false); ok {
				setMeta(out, "img dpi", trimFloat(v))
			}
		case 0x8769:
			exifOffset = tr.entryUint(e)
		case 0x8825:
			gpsOffset = tr.entryUint(e)
		}
	}

	if exifOffset > 0 {
		for _, e := range tr.readIFD(exifOffset) {
			switch e.tag {
			case 0x9003:
				setMeta(out, "img dateoriginal", entryString(e))
			case 0x9004:
				setMeta(out, "img datecreate", entryString(e))
			case 0x9291:
				setMeta(out, "img subsec", entryString(e))
			case 0xa002:
				setMeta(out, "width", strconv.FormatUint(uint64(tr.entryUint(e)), 10))
			case 0xa003:
				setMeta(out, "height", strconv.FormatUint(uint64(tr.entryUint(e)), 10))
			}
		}
	}

	if date := out["img dateoriginal"]; date != "" {
		out["img timeoriginal"] = date
	}
	if date := out["img datecreate"]; date != "" {
		out["img timecreate"] = date
	}

	if gpsOffset > 0 {
		readGPSMetadata(tr, gpsOffset, out)
	}

	return out
}

func extractJPEGExif(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var soi [2]byte
	if _, err := io.ReadFull(f, soi[:]); err != nil || soi != [2]byte{0xff, 0xd8} {
		return nil
	}

	for {
		var prefix [1]byte
		if _, err := io.ReadFull(f, prefix[:]); err != nil {
			return nil
		}
		if prefix[0] != 0xff {
			continue
		}
		var marker [1]byte
		for {
			if _, err := io.ReadFull(f, marker[:]); err != nil {
				return nil
			}
			if marker[0] != 0xff {
				break
			}
		}
		if marker[0] == 0xd9 || marker[0] == 0xda {
			return nil
		}
		var lenBuf [2]byte
		if _, err := io.ReadFull(f, lenBuf[:]); err != nil {
			return nil
		}
		n := int(binary.BigEndian.Uint16(lenBuf[:]))
		if n < 2 || n > 32*1024*1024 {
			return nil
		}
		payload := make([]byte, n-2)
		if _, err := io.ReadFull(f, payload); err != nil {
			return nil
		}
		if marker[0] == 0xe1 && len(payload) > 6 && bytes.Equal(payload[:6], []byte{'E', 'x', 'i', 'f', 0, 0}) {
			return payload[6:]
		}
	}
}

func newTIFFReader(data []byte) (*tiffReader, bool) {
	if len(data) < 8 {
		return nil, false
	}
	var order binary.ByteOrder
	switch string(data[:2]) {
	case "II":
		order = binary.LittleEndian
	case "MM":
		order = binary.BigEndian
	default:
		return nil, false
	}
	if order.Uint16(data[2:4]) != 42 {
		return nil, false
	}
	return &tiffReader{data: data, order: order}, true
}

func (t *tiffReader) u16(off uint32) uint16 {
	if uint64(off)+2 > uint64(len(t.data)) {
		return 0
	}
	return t.order.Uint16(t.data[off : off+2])
}

func (t *tiffReader) u32(off uint32) uint32 {
	if uint64(off)+4 > uint64(len(t.data)) {
		return 0
	}
	return t.order.Uint32(t.data[off : off+4])
}

func (t *tiffReader) readIFD(off uint32) []tiffEntry {
	if off == 0 || uint64(off)+2 > uint64(len(t.data)) {
		return nil
	}
	count := int(t.u16(off))
	if count > 4096 {
		return nil
	}
	entries := make([]tiffEntry, 0, count)
	base := off + 2
	for i := 0; i < count; i++ {
		pos := base + uint32(i*12)
		if uint64(pos)+12 > uint64(len(t.data)) {
			break
		}
		tag := t.u16(pos)
		typ := t.u16(pos + 2)
		n := t.u32(pos + 4)
		size := typeSize(typ)
		if size == 0 || n == 0 || uint64(size)*uint64(n) > 64*1024*1024 {
			continue
		}
		total := uint32(size) * n
		var raw []byte
		if total <= 4 {
			raw = append([]byte(nil), t.data[pos+8:pos+8+total]...)
		} else {
			valueOff := t.u32(pos + 8)
			if uint64(valueOff)+uint64(total) > uint64(len(t.data)) {
				continue
			}
			raw = append([]byte(nil), t.data[valueOff:valueOff+total]...)
		}
		entries = append(entries, tiffEntry{tag: tag, typ: typ, count: n, raw: raw})
	}
	return entries
}

func typeSize(typ uint16) uint32 {
	switch typ {
	case 1, 2, 6, 7:
		return 1
	case 3, 8:
		return 2
	case 4, 9, 11:
		return 4
	case 5, 10, 12:
		return 8
	default:
		return 0
	}
}

func (t *tiffReader) entryUint(e tiffEntry) uint32 {
	switch e.typ {
	case 1, 7:
		if len(e.raw) > 0 {
			return uint32(e.raw[0])
		}
	case 3:
		if len(e.raw) >= 2 {
			return uint32(t.order.Uint16(e.raw[:2]))
		}
	case 4:
		if len(e.raw) >= 4 {
			return t.order.Uint32(e.raw[:4])
		}
	}
	return 0
}

func (t *tiffReader) entryRational(e tiffEntry, signed bool) (float64, bool) {
	if len(e.raw) < 8 {
		return 0, false
	}
	if signed || e.typ == 10 {
		num := int32(t.order.Uint32(e.raw[:4]))
		den := int32(t.order.Uint32(e.raw[4:8]))
		if den == 0 {
			return 0, false
		}
		return float64(num) / float64(den), true
	}
	num := t.order.Uint32(e.raw[:4])
	den := t.order.Uint32(e.raw[4:8])
	if den == 0 {
		return 0, false
	}
	return float64(num) / float64(den), true
}

func entryString(e tiffEntry) string {
	if len(e.raw) == 0 {
		return ""
	}
	return strings.TrimSpace(strings.TrimRight(string(e.raw), "\x00"))
}

func decodeXPString(raw []byte) string {
	if len(raw) < 2 {
		return ""
	}
	u := make([]uint16, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		v := binary.LittleEndian.Uint16(raw[i : i+2])
		if v == 0 {
			break
		}
		u = append(u, v)
	}
	return strings.TrimSpace(string(utf16.Decode(u)))
}

func readGPSMetadata(t *tiffReader, off uint32, out map[string]string) {
	var latRef, lngRef string
	var lat, lng [3]float64
	var latOK, lngOK bool
	alt := math.NaN()
	altRef := uint32(0)

	for _, e := range t.readIFD(off) {
		switch e.tag {
		case 1:
			latRef = strings.ToUpper(entryString(e))
		case 2:
			if values, ok := t.rationalArray(e, 3); ok {
				copy(lat[:], values)
				latOK = true
			}
		case 3:
			lngRef = strings.ToUpper(entryString(e))
		case 4:
			if values, ok := t.rationalArray(e, 3); ok {
				copy(lng[:], values)
				lngOK = true
			}
		case 5:
			altRef = t.entryUint(e)
		case 6:
			if value, ok := t.entryRational(e, false); ok {
				alt = value
			}
		}
	}

	if latOK {
		decimal := lat[0] + lat[1]/60 + lat[2]/3600
		if latRef == "S" {
			decimal = -decimal
		}
		setMeta(out, "gps lat", fmt.Sprintf("%.6f", decimal))
		setMeta(out, "gps lat deg", trimFloat(lat[0]))
		setMeta(out, "gps lat min", trimFloat(lat[1]))
		setMeta(out, "gps lat sec", trimFloat(lat[2]))
		setMeta(out, "gps lat dir", latRef)
	}
	if lngOK {
		decimal := lng[0] + lng[1]/60 + lng[2]/3600
		if lngRef == "W" {
			decimal = -decimal
		}
		setMeta(out, "gps lng", fmt.Sprintf("%.6f", decimal))
		setMeta(out, "gps lng deg", trimFloat(lng[0]))
		setMeta(out, "gps lng min", trimFloat(lng[1]))
		setMeta(out, "gps lng sec", trimFloat(lng[2]))
		setMeta(out, "gps lng dir", lngRef)
	}
	if !math.IsNaN(alt) {
		if altRef == 1 {
			alt = -alt
		}
		setMeta(out, "gps alt", trimFloat(alt))
	}
}

func (t *tiffReader) rationalArray(e tiffEntry, count int) ([]float64, bool) {
	if e.typ != 5 && e.typ != 10 {
		return nil, false
	}
	if len(e.raw) < count*8 {
		return nil, false
	}
	values := make([]float64, 0, count)
	for i := 0; i < count; i++ {
		part := e.raw[i*8 : i*8+8]
		tmp := tiffEntry{typ: e.typ, raw: part}
		v, ok := t.entryRational(tmp, e.typ == 10)
		if !ok {
			return nil, false
		}
		values = append(values, v)
	}
	return values, true
}

func setMeta(m map[string]string, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	m[strings.ToLower(key)] = value
}
