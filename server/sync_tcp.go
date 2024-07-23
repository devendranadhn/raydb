package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"ray/config"
	"ray/core"
	"strconv"
	"strings"
)

func RunSyncTcpServer() {
	log.Println("starting a synchronous TCP server on ", config.Host, config.Port)

	var con_clients int = 0

	//listening to the configure port
	listener, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))

	if err != nil {
		panic(err)
	}

	for {
		conn, err := listener.Accept()

		if err != nil {
			panic(err)
		}

		//increment the no of concurrent clients
		con_clients += 1

		log.Println("client connected with address:", conn.RemoteAddr(), " concurrent clients : ", con_clients)

		for {
			// over the socket continously read the command and print it out.
			cmds, err := readCommands(conn)
			if err != nil {
				conn.Close()
				con_clients -= 1
				log.Println("client disconnected : ", conn.RemoteAddr(), " concurrent clients : ", con_clients)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}

			log.Println("command : ", cmds)

			respond(cmds, conn)

		}

	}

}

func readCommands(conn io.ReadWriter) (core.RedisCmds, error) {

	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:])

	if err != nil {
		return nil, err
	}

	values, err := core.Decode(buf[:n])
	if err != nil {
		return nil, err
	}

	var cmds []*core.RedisCmd = make([]*core.RedisCmd, 0)

	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}

		cmds = append(cmds, &core.RedisCmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		})
	}

	return cmds, nil

}

func toArrayString(ai []interface{}) ([]string, error) {
	as := make([]string, len(ai))
	for i := range ai {
		as[i] = ai[i].(string)
	}
	return as, nil
}

func respond(cmds core.RedisCmds, conn io.ReadWriter) {

	err := core.EvalAndRespond(cmds, conn)

	if err != nil {
		respondError(err, conn)
	}
}

func respondError(err error, conn io.ReadWriter) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}
