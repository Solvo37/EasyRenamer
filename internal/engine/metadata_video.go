package engine

import (
	"encoding/binary"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type mp4Atom struct {
	typ       string
	offset    int64
	size      int64
	header    int64
	dataStart int64
	dataSize  int64
}

func readVideoMetadata(path string) map[string]string {
	out := map[string]string{}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp4", ".m4v", ".mov", ".3gp":
	default:
		return out
	}

	f, err := os.Open(path)
	if err != nil {
		return out
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return out
	}

	top := readMP4Atoms(f, 0, st.Size(), 128)
	var moov mp4Atom
	for _, atom := range top {
		if atom.typ == "moov" {
			moov = atom
			break
		}
	}
	if moov.size == 0 {
		return out
	}

	children := readMP4Atoms(f, moov.dataStart, moov.dataSize, 4096)
	for _, atom := range children {
		if atom.typ == "mvhd" {
			if created, _, _, ok := parseMVHD(f, atom); ok {
				if !created.IsZero() {
					setMeta(out, "video date", created.Format("2006-01-02 15:04:05"))
				}
			}
		}
	}

	for _, atom := range children {
		if atom.typ != "trak" {
			continue
		}
		track := parseVideoTrack(f, atom)
		if !track.video {
			continue
		}
		if track.width > 0 {
			setMeta(out, "width", strconv.Itoa(track.width))
		}
		if track.height > 0 {
			setMeta(out, "height", strconv.Itoa(track.height))
		}
		if track.duration > 0 {
			setMeta(out, "duration seconds", trimFloat(track.duration))
		}
		if track.frameRate > 0 {
			setMeta(out, "framerate", trimFloat(track.frameRate))
		}
		break
	}

	readMP4TextMetadata(f, moov, out)
	return out
}

type mp4TrackInfo struct {
	video     bool
	width     int
	height    int
	duration  float64
	frameRate float64
}

func parseVideoTrack(r io.ReaderAt, atom mp4Atom) mp4TrackInfo {
	var info mp4TrackInfo
	children := readMP4Atoms(r, atom.dataStart, atom.dataSize, 4096)
	for _, child := range children {
		switch child.typ {
		case "tkhd":
			info.width, info.height = parseTKHD(r, child)
		case "mdia":
			media := parseMP4Media(r, child)
			if media.handler == "vide" {
				info.video = true
				info.duration = media.duration
				info.frameRate = media.frameRate
			}
		}
	}
	return info
}

type mp4MediaInfo struct {
	handler   string
	duration  float64
	frameRate float64
}

func parseMP4Media(r io.ReaderAt, atom mp4Atom) mp4MediaInfo {
	var out mp4MediaInfo
	var timescale uint32
	children := readMP4Atoms(r, atom.dataStart, atom.dataSize, 4096)
	for _, child := range children {
		switch child.typ {
		case "hdlr":
			buf := readAtMost(r, child.dataStart, min64(child.dataSize, 32))
			if len(buf) >= 12 {
				out.handler = string(buf[8:12])
			}
		case "mdhd":
			_, timescale, out.duration, _ = parseMDHD(r, child)
		case "minf":
			if timescale == 0 {
				continue
			}
			samples, mediaDuration := findSTTSTotals(r, child)
			if samples > 0 && mediaDuration > 0 {
				seconds := float64(mediaDuration) / float64(timescale)
				if seconds > 0 {
					out.frameRate = float64(samples) / seconds
				}
			}
		}
	}
	if out.frameRate > 1000 || math.IsNaN(out.frameRate) || math.IsInf(out.frameRate, 0) {
		out.frameRate = 0
	}
	return out
}

func findSTTSTotals(r io.ReaderAt, atom mp4Atom) (uint64, uint64) {
	children := readMP4Atoms(r, atom.dataStart, atom.dataSize, 4096)
	for _, child := range children {
		if child.typ == "stts" {
			buf := readAtMost(r, child.dataStart, child.dataSize)
			if len(buf) < 8 {
				return 0, 0
			}
			count := int(binary.BigEndian.Uint32(buf[4:8]))
			if count > 100000 || 8+count*8 > len(buf) {
				return 0, 0
			}
			var samples, duration uint64
			for i := 0; i < count; i++ {
				off := 8 + i*8
				n := uint64(binary.BigEndian.Uint32(buf[off : off+4]))
				delta := uint64(binary.BigEndian.Uint32(buf[off+4 : off+8]))
				samples += n
				duration += n * delta
			}
			return samples, duration
		}
		if child.typ == "stbl" || child.typ == "minf" {
			if s, d := findSTTSTotals(r, child); s > 0 {
				return s, d
			}
		}
	}
	return 0, 0
}

func parseTKHD(r io.ReaderAt, atom mp4Atom) (int, int) {
	buf := readAtMost(r, atom.dataStart, min64(atom.dataSize, 128))
	if len(buf) < 84 {
		return 0, 0
	}
	version := buf[0]
	off := 76
	if version == 1 {
		off = 88
	}
	if off+8 > len(buf) {
		return 0, 0
	}
	width := int(binary.BigEndian.Uint32(buf[off:off+4]) >> 16)
	height := int(binary.BigEndian.Uint32(buf[off+4:off+8]) >> 16)
	return width, height
}

func parseMVHD(r io.ReaderAt, atom mp4Atom) (time.Time, uint32, float64, bool) {
	return parseTimeHeader(r, atom)
}

func parseMDHD(r io.ReaderAt, atom mp4Atom) (time.Time, uint32, float64, bool) {
	return parseTimeHeader(r, atom)
}

func parseTimeHeader(r io.ReaderAt, atom mp4Atom) (time.Time, uint32, float64, bool) {
	buf := readAtMost(r, atom.dataStart, min64(atom.dataSize, 48))
	if len(buf) < 20 {
		return time.Time{}, 0, 0, false
	}
	version := buf[0]
	var created uint64
	var timescale uint32
	var duration uint64
	if version == 1 {
		if len(buf) < 32 {
			return time.Time{}, 0, 0, false
		}
		created = binary.BigEndian.Uint64(buf[4:12])
		timescale = binary.BigEndian.Uint32(buf[20:24])
		duration = binary.BigEndian.Uint64(buf[24:32])
	} else {
		created = uint64(binary.BigEndian.Uint32(buf[4:8]))
		timescale = binary.BigEndian.Uint32(buf[12:16])
		duration = uint64(binary.BigEndian.Uint32(buf[16:20]))
	}
	var t time.Time
	if created > 0 {
		const unixOffset = 2082844800
		if created > unixOffset {
			t = time.Unix(int64(created-unixOffset), 0)
		}
	}
	seconds := 0.0
	if timescale > 0 {
		seconds = float64(duration) / float64(timescale)
	}
	return t, timescale, seconds, true
}

func readMP4TextMetadata(r io.ReaderAt, moov mp4Atom, out map[string]string) {
	var visit func(mp4Atom, int)
	visit = func(parent mp4Atom, depth int) {
		if depth > 8 {
			return
		}
		start := parent.dataStart
		size := parent.dataSize
		if parent.typ == "meta" {
			start += 4
			size -= 4
			if size <= 0 {
				return
			}
		}
		for _, child := range readMP4Atoms(r, start, size, 4096) {
			if parent.typ == "ilst" {
				switch child.typ {
				case "\xa9nam":
					if v := readMP4DataString(r, child); v != "" {
						setMetaIfEmpty(out, "title", v)
					}
				case "\xa9gen":
					if v := readMP4DataString(r, child); v != "" {
						setMetaIfEmpty(out, "genre", v)
					}
				case "\xa9day":
					if v := readMP4DataString(r, child); v != "" {
						setMetaIfEmpty(out, "video date", v)
					}
				}
			}
			switch child.typ {
			case "udta", "meta", "ilst":
				visit(child, depth+1)
			}
		}
	}
	visit(moov, 0)
}

func readMP4DataString(r io.ReaderAt, atom mp4Atom) string {
	for _, child := range readMP4Atoms(r, atom.dataStart, atom.dataSize, 64) {
		if child.typ != "data" || child.dataSize <= 8 {
			continue
		}
		buf := readAtMost(r, child.dataStart+8, child.dataSize-8)
		return strings.TrimSpace(strings.TrimRight(string(buf), "\x00"))
	}
	return ""
}

func readMP4Atoms(r io.ReaderAt, start, length int64, maxAtoms int) []mp4Atom {
	if length <= 0 {
		return nil
	}
	end := start + length
	pos := start
	atoms := make([]mp4Atom, 0)
	for pos+8 <= end && len(atoms) < maxAtoms {
		header := readAtMost(r, pos, min64(16, end-pos))
		if len(header) < 8 {
			break
		}
		size := int64(binary.BigEndian.Uint32(header[0:4]))
		typ := string(header[4:8])
		headerSize := int64(8)
		if size == 1 {
			if len(header) < 16 {
				break
			}
			size = int64(binary.BigEndian.Uint64(header[8:16]))
			headerSize = 16
		} else if size == 0 {
			size = end - pos
		}
		if size < headerSize || pos+size > end {
			break
		}
		atoms = append(atoms, mp4Atom{
			typ: typ, offset: pos, size: size, header: headerSize,
			dataStart: pos + headerSize, dataSize: size - headerSize,
		})
		pos += size
	}
	return atoms
}

func readAtMost(r io.ReaderAt, off, n int64) []byte {
	if n <= 0 || n > 64*1024*1024 {
		return nil
	}
	buf := make([]byte, int(n))
	read, err := r.ReadAt(buf, off)
	if err != nil && err != io.EOF {
		return nil
	}
	return buf[:read]
}

func min64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
