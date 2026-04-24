package encoding

type Encoding int

const (
	EncodingAuto Encoding = iota
	EncodingUTF8
	EncodingUTF16LE
	EncodingUTF16BE
	EncodingISO88591
	EncodingWindows1252
)

func (e Encoding) String() string {
	switch e {
	case EncodingAuto:
		return "auto"
	case EncodingUTF8:
		return "utf-8"
	case EncodingUTF16LE:
		return "utf-16le"
	case EncodingUTF16BE:
		return "utf-16be"
	case EncodingISO88591:
		return "iso-8859-1"
	case EncodingWindows1252:
		return "windows-1252"
	default:
		return "unknown"
	}
}

func DetectBOM(data []byte) (Encoding, int) {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return EncodingUTF8, 3
	}
	if len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE {
		return EncodingUTF16LE, 2
	}
	if len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF {
		return EncodingUTF16BE, 2
	}
	return EncodingUTF8, 0
}
