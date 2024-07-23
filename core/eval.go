package core

import (
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

// func EvalAndRespond(cmd *RedisCmd, conn net.Conn) error {
// 	switch cmd.Cmd {
// 	case "PING":
// 		return evalPing(cmd.Args, conn)
// 	default:
// 		return evalPing(cmd.Args, conn)
// 	}
// }

func EvalAndRespond(cmd *RedisCmd, conn io.ReadWriter) error {
	switch cmd.Cmd {
	case "PING":
		return evalPing(cmd.Args, conn)
	case "SET":
		return evalSET(cmd.Args, conn)
	case "GET":
		return evalGET(cmd.Args, conn)
	case "TTL":
		return evalTTL(cmd.Args, conn)
	case "DELETE":
		return evalDELETE(cmd.Args, conn)
	case "EXPIRE":
		return evalEXPIRE(cmd.Args, conn)
	default:
		return evalPing(cmd.Args, conn)
	}
}

// func evalPing(args []string, conn net.Conn) error {
// 	var b []byte

// 	if len(args) >= 2 {
// 		return errors.New("ERR wrong number of commands for 'ping' command")
// 	}
// 	if len(args) == 0 {
// 		b = Encode("PONG", true)
// 	} else {
// 		b = Encode(args[0], false)
// 	}

// 	_, err := conn.Write(b)
// 	return err
// }

func evalPing(args []string, conn io.ReadWriter) error {
	var b []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of commands for 'ping' command")
	}
	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := conn.Write(b)
	return err
}

func evalSET(args []string, conn io.ReadWriter) error {
	if len(args) < 1 {
		return errors.New("(err) wrong number of arguments for the 'set' command")
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
					return errors.New("(error) ERR syntax error")
				}
				var expirationSecStr = args[i]
				expirationSec, err := strconv.ParseInt(expirationSecStr, 10, 64)
				if err != nil {
					return errors.New("(ERR) value is not an integer or out of range")
				}
				expirationMs = expirationSec * 1000
			}
		}
	}
	Put(key, NewObj(value, expirationMs))
	// log.Println("received set command with key:", key, " value:", value)
	conn.Write(RESP_OK)
	return nil

}

func evalGET(args []string, conn io.ReadWriter) error {

	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for the 'get' command")
	}

	var key string = args[0]

	obj := Get(key)

	// if key doesn't present then return nil
	if obj == nil {
		conn.Write(RESP_NIL)
		return nil
	}

	// if key present and expired, then clear the key and return nil
	if obj.ExpiresAt != -1 && obj.ExpiresAt <= time.Now().UnixMilli() {
		conn.Write(RESP_NIL)
		return nil
	}

	// return RESP encoded value
	conn.Write(Encode(obj.Value, false))

	return nil
}

func evalTTL(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("(error) ERR wrong number of arguments for 'ttl' command")
	}
	var key string = args[0]

	obj := Get(key)

	// if key doesn't exist return RESP encoded -2 denoting key doesn't exist
	if obj == nil {
		conn.Write(RESP_MINUS_2)
		return nil
	}

	// if object exist but no expiration is set, then return RESP encoded -1 denoting expiry not set
	if obj.ExpiresAt == -1 {
		conn.Write(RESP_MINUS_1)
		return nil
	}

	// compute the time remaining for expiry
	// return RESP encoded time
	durationMs := obj.ExpiresAt - time.Now().UnixMilli()
	if durationMs < 0 {
		conn.Write(RESP_MINUS_2)
		return nil
	}

	conn.Write(Encode(int64(durationMs/1000), false))

	return nil
}

func evalDELETE(args []string, conn io.ReadWriter) error {

	var countDeleted int = 0
	for _, key := range args {
		if ok := Delete(key); ok {
			countDeleted++
		}
	}
	conn.Write(Encode(countDeleted, false))
	return nil

}

func evalEXPIRE(args []string, conn io.ReadWriter) error {

	if len(args) <= 1 {
		return errors.New("(error) ERR wrong number of arguments for 'expire' command")
	}

	var key string = args[0]

	exDurationSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return errors.New("(error) ERR value is not an integer or not in range")
	}

	obj := Get(key)

	if obj == nil {
		conn.Write(RESP_ZERO)
		return nil
	}

	// update the expiresat for the obj
	obj.ExpiresAt = time.Now().UnixMilli() + exDurationSec*1000

	conn.Write(RESP_ONE)

	return nil

}

func Encode(value interface{}, isSimple bool) []byte {

	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		} else {
			return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
		}
	case int, int8, int16, int32, int64:
		return []byte(fmt.Sprintf(":%d\r\n", value))
	}

	return []byte{}
}
