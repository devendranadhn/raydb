package core

import (
	"errors"
)

func Decode(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	var values []interface{} = make([]interface{}, 0)

	var index int = 0

	for index < len(data) {
		value, delta, err := DecodeOne(data[index:])
		if err != nil {
			return values, err
		}
		index = index + delta
		values = append(values, value)
	}
	return values, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {

	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '-':
		return readError(data)
	case ':':
		return readInteger(data)
	case '$':
		return readBulkString(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil

}

// reads the RESP encoded simple string from data and returns
// the string, the delta, the error
func readSimpleString(data []byte) (string, int, error) {
	pos := 1
	for ; data[pos] != '\r'; pos++ {
	}
	return string(data[1:pos]), pos + 2, nil
}

// reads the RESP encoded simple string from data and returns
// the string, the delta, the error
func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

// reads the RESP encoded integer from data and returns
// the string, the delta, the error
func readInteger(data []byte) (int64, int, error) {
	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}
	return value, pos + 2, nil
}

// reads the RESP bulk string from data and returns
// the string, the delta, the error
func readBulkString(data []byte) (string, int, error) {
	pos := 1

	//reading the length and forwarding the pos by
	// the length of the integer + the first special character
	len, delta := readLength(data[pos:])

	pos = pos + delta

	return string(data[pos:(pos + len)]), pos + len + 2, nil

}

// reads the length typically the first integer of the string
// until it hits by a non digit byte and returns
// the integer and the delta = length + 2(CRLF)

func readLength(data []byte) (int, int) {
	pos, length := 0, 0
	for pos = range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}
	return 0, 0
}

func readArray(data []byte) (interface{}, int, error) {
	pos := 1

	count, delta := readLength(data[pos:])

	pos = pos + delta

	var elems []interface{} = make([]interface{}, count)

	for i := range elems {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elems[i] = elem
		pos = pos + delta
	}
	return elems, pos, nil
}
