package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

var RESP_NIL []byte = []byte("$-1\r\n")
var RESP_OK []byte = []byte("+OK\r\n")
var RESP_QUEUED []byte = []byte("+QUEUED\r\n")
var RESP_ZERO []byte = []byte(":0\r\n")
var RESP_ONE []byte = []byte(":1\r\n")
var RESP_MINUS_1 []byte = []byte(":-1\r\n")
var RESP_MINUS_2 []byte = []byte(":-2\r\n")
var RESP_EMPTY_ARRAY []byte = []byte("*0\r\n")

func EvalAndRespond(cmds RedisCmds, conn io.ReadWriter) error {

	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {

		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPING(cmd.Args))
		case "SET":
			buf.Write(evalSET(cmd.Args))
		case "GET":
			buf.Write(evalGET(cmd.Args))
		case "TTL":
			buf.Write(evalTTL(cmd.Args))
		case "DELETE":
			buf.Write(evalDELETE(cmd.Args))
		case "EXPIRE":
			buf.Write(evalEXPIRE(cmd.Args))
		case "BGREWRITEAOF":
			buf.Write(evalBGREWRITEAOF(cmd.Args))
		default:
			buf.Write(evalPING(cmd.Args))
		}
	}
	conn.Write(buf.Bytes())

	return nil
}

func evalPING(args []string) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("ERR wrong number of commands for 'ping' command"), false)
	}
	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}
	return b
}

func evalSET(args []string) []byte {
	if len(args) < 1 {
		return Encode(errors.New("(err) wrong number of arguments for the 'set' command"), false)
	}

	var key, value string
	var expirationMs int64 = -1

	key, value = args[0], args[1]

	// log.Println("received set command with key:", key, " value:", value)

	for i := 2; i < len(args); i++ {
		switch args[2] {
		case "EX", "ex":
			i++
			{
				if i == len(args) {
					return Encode(errors.New("(error) ERR syntax error"), false)
				}
				var expirationSecStr = args[i]
				expirationSec, err := strconv.ParseInt(expirationSecStr, 10, 64)
				if err != nil {
					return Encode(errors.New("(ERR) value is not an integer or out of range"), false)
				}
				expirationMs = expirationSec * 1000
			}
		}
	}
	Put(key, NewObj(value, expirationMs))
	// log.Println("received set command with key:", key, " value:", value)
	return RESP_OK

}

func evalGET(args []string) []byte {

	if len(args) != 1 {
		return Encode(errors.New("(error) ERR wrong number of arguments for the 'get' command"), false)
	}

	var key string = args[0]

	obj := Get(key)

	// if key doesn't present then return nil
	if obj == nil {
		return RESP_NIL
	}

	// if key present and expired, then clear the key and return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		return RESP_NIL
	}

	// return RESP encoded value
	return Encode(obj.Value, false)
}

func evalTTL(args []string) []byte {
	if len(args) != 1 {
		Encode(errors.New("(error) ERR wrong number of arguments for 'ttl' command"), false)
	}
	var key string = args[0]

	obj := Get(key)

	// if key doesn't exist return RESP encoded -2 denoting key doesn't exist
	if obj == nil {
		return RESP_MINUS_2
	}

	// if object exist but no expiration is set, then return RESP encoded -1 denoting expiry not set
	if obj.ExpiresAt == -1 {
		return RESP_MINUS_1
	}

	// compute the time remaining for expiry
	// return RESP encoded time
	durationMs := obj.ExpiresAt - time.Now().UnixMilli()
	if durationMs < 0 {
		return RESP_MINUS_2
	}

	return Encode(int64(durationMs/1000), false)
}

func evalDELETE(args []string) []byte {

	var countDeleted int = 0
	for _, key := range args {
		if ok := Delete(key); ok {
			countDeleted++
		}
	}
	return Encode(countDeleted, false)

}

func evalEXPIRE(args []string) []byte {

	if len(args) <= 1 {
		Encode(errors.New("(error) ERR wrong number of arguments for 'expire' command"), false)
	}

	var key string = args[0]

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("(error) ERR value is not an integer or not in range"), false)
	}

	obj := Get(key)

	if obj == nil {
		return RESP_ZERO
	}

	// update the expiresat for the obj
	obj.ExpiresAt = time.Now().UnixMilli() + exDurationSec*1000

	return RESP_ONE
}

func evalBGREWRITEAOF(args []string) []byte {
	DumpAllAOF()
	return RESP_OK
}

func encodeString(v string) []byte {
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
}

func Encode(value interface{}, isSimple bool) []byte {

	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		} else {
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		}
	case []string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, b := range value.([]string) {
			buf.Write(encodeString(b))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(v), buf.Bytes()))
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", value))
	}

	return []byte{}
}
