package engine

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf16"
)

func readAudioMetadata(path string) map[string]string {
	out := map[string]string{}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp3":
		readID3v2(path, out)
		readID3v1(path, out)
	case ".flac":
		readFLACComments(path, out)
	}
	return out
}

func readID3v2(path string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	header := make([]byte, 10)
	if _, err := io.ReadFull(f, header); err != nil || string(header[:3]) != "ID3" {
		return
	}
	major := header[3]
	if major < 3 || major > 4 {
		return
	}
	size := synchsafe32(header[6:10])
	if size <= 0 || size > 32*1024*1024 {
		return
	}
	data := make([]byte, size)
	if _, err := io.ReadFull(f, data); err != nil {
		return
	}

	pos := 0
	for pos+10 <= len(data) {
		id := string(data[pos : pos+4])
		if strings.Trim(id, "\x00") == "" {
			break
		}
		var frameSize int
		if major == 4 {
			frameSize = synchsafe32(data[pos+4 : pos+8])
		} else {
			frameSize = int(binary.BigEndian.Uint32(data[pos+4 : pos+8]))
		}
		pos += 10
		if frameSize <= 0 || pos+frameSize > len(data) {
			break
		}
		payload := data[pos : pos+frameSize]
		pos += frameSize

		switch id {
		case "TIT2":
			setMetaIfEmpty(out, "title", decodeID3Text(payload))
		case "TPE1":
			setMetaIfEmpty(out, "artist", decodeID3Text(payload))
		case "TALB":
			setMetaIfEmpty(out, "album", decodeID3Text(payload))
		case "TCON":
			setMetaIfEmpty(out, "genre", decodeID3Text(payload))
		case "TYER", "TDRC":
			year := decodeID3Text(payload)
			if len(year) >= 4 {
				year = year[:4]
			}
			setMetaIfEmpty(out, "audio year", year)
		case "TRCK":
			part, total := splitNumberPair(decodeID3Text(payload))
			setMetaIfEmpty(out, "track", part)
			setMetaIfEmpty(out, "trackcount", total)
		case "TPOS":
			part, total := splitNumberPair(decodeID3Text(payload))
			setMetaIfEmpty(out, "disc", part)
			setMetaIfEmpty(out, "disccount", total)
		}
	}
}

func readID3v1(path string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil || st.Size() < 128 {
		return
	}
	if _, err := f.Seek(-128, io.SeekEnd); err != nil {
		return
	}
	buf := make([]byte, 128)
	if _, err := io.ReadFull(f, buf); err != nil || string(buf[:3]) != "TAG" {
		return
	}

	setMetaIfEmpty(out, "title", latin1Trim(buf[3:33]))
	setMetaIfEmpty(out, "artist", latin1Trim(buf[33:63]))
	setMetaIfEmpty(out, "album", latin1Trim(buf[63:93]))
	setMetaIfEmpty(out, "audio year", latin1Trim(buf[93:97]))
	if buf[125] == 0 && buf[126] != 0 {
		setMetaIfEmpty(out, "track", strconv.Itoa(int(buf[126])))
	}
}

func synchsafe32(b []byte) int {
	if len(b) < 4 {
		return 0
	}
	return int(b[0]&0x7f)<<21 | int(b[1]&0x7f)<<14 | int(b[2]&0x7f)<<7 | int(b[3]&0x7f)
}

func decodeID3Text(payload []byte) string {
	if len(payload) == 0 {
		return ""
	}
	enc := payload[0]
	data := payload[1:]
	switch enc {
	case 0:
		return latin1Trim(data)
	case 1:
		return decodeUTF16WithBOM(data)
	case 2:
		return decodeUTF16BE(data)
	case 3:
		return strings.TrimSpace(strings.TrimRight(string(data), "\x00"))
	default:
		return ""
	}
}

func latin1Trim(data []byte) string {
	runes := make([]rune, 0, len(data))
	for _, b := range data {
		if b == 0 {
			break
		}
		runes = append(runes, rune(b))
	}
	return strings.TrimSpace(string(runes))
}

func decodeUTF16WithBOM(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	if data[0] == 0xff && data[1] == 0xfe {
		return decodeUTF16(data[2:], binary.LittleEndian)
	}
	if data[0] == 0xfe && data[1] == 0xff {
		return decodeUTF16(data[2:], binary.BigEndian)
	}
	return decodeUTF16(data, binary.LittleEndian)
}

func decodeUTF16BE(data []byte) string {
	return decodeUTF16(data, binary.BigEndian)
}

func decodeUTF16(data []byte, order binary.ByteOrder) string {
	u := make([]uint16, 0, len(data)/2)
	for i := 0; i+1 < len(data); i += 2 {
		v := order.Uint16(data[i : i+2])
		if v == 0 {
			break
		}
		u = append(u, v)
	}
	return strings.TrimSpace(string(utf16.Decode(u)))
}

func splitNumberPair(value string) (string, string) {
	value = strings.TrimSpace(value)
	parts := strings.SplitN(value, "/", 2)
	first := strings.TrimSpace(parts[0])
	second := ""
	if len(parts) == 2 {
		second = strings.TrimSpace(parts[1])
	}
	return first, second
}

func setMetaIfEmpty(out map[string]string, key, value string) {
	key = strings.ToLower(key)
	if out[key] != "" {
		return
	}
	setMeta(out, key, value)
}

func readFLACComments(path string, out map[string]string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	var magic [4]byte
	if _, err := io.ReadFull(f, magic[:]); err != nil || string(magic[:]) != "fLaC" {
		return
	}

	for block := 0; block < 128; block++ {
		var h [4]byte
		if _, err := io.ReadFull(f, h[:]); err != nil {
			return
		}
		last := h[0]&0x80 != 0
		typ := h[0] & 0x7f
		length := int(h[1])<<16 | int(h[2])<<8 | int(h[3])
		if length < 0 || length > 32*1024*1024 {
			return
		}
		data := make([]byte, length)
		if _, err := io.ReadFull(f, data); err != nil {
			return
		}
		if typ == 4 {
			parseVorbisComments(data, out)
		}
		if last {
			return
		}
	}
}

func parseVorbisComments(data []byte, out map[string]string) {
	r := bytes.NewReader(data)
	var vendorLen uint32
	if err := binary.Read(r, binary.LittleEndian, &vendorLen); err != nil || vendorLen > uint32(r.Len()) {
		return
	}
	if _, err := r.Seek(int64(vendorLen), io.SeekCurrent); err != nil {
		return
	}
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil || count > 100000 {
		return
	}
	for i := uint32(0); i < count; i++ {
		var n uint32
		if err := binary.Read(r, binary.LittleEndian, &n); err != nil || n > uint32(r.Len()) {
			return
		}
		buf := make([]byte, n)
		if _, err := io.ReadFull(r, buf); err != nil {
			return
		}
		pair := strings.SplitN(string(buf), "=", 2)
		if len(pair) != 2 {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(pair[0]))
		value := strings.TrimSpace(pair[1])
		switch key {
		case "TITLE":
			setMetaIfEmpty(out, "title", value)
		case "ARTIST", "ALBUMARTIST":
			setMetaIfEmpty(out, "artist", value)
		case "ALBUM":
			setMetaIfEmpty(out, "album", value)
		case "GENRE":
			setMetaIfEmpty(out, "genre", value)
		case "DATE", "YEAR":
			if len(value) >= 4 {
				value = value[:4]
			}
			setMetaIfEmpty(out, "audio year", value)
		case "TRACKNUMBER":
			setMetaIfEmpty(out, "track", value)
		case "TRACKTOTAL", "TOTALTRACKS":
			setMetaIfEmpty(out, "trackcount", value)
		case "DISCNUMBER":
			setMetaIfEmpty(out, "disc", value)
		case "DISCTOTAL", "TOTALDISCS":
			setMetaIfEmpty(out, "disccount", value)
		}
	}
}
