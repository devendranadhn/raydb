package core

import "strconv"

func deduceTypeEncoding(v string) (uint8, uint8) {
	oType := OBJ_ENCODING_EMBSTR
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return oType, OBJ_ENCODING_INT
	}
	if len(v) <= 44 {
		return oType, OBJ_ENCODING_EMBSTR
	}
	return oType, OBJ_ENCODING_RAW
}
