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
			cmd, err := readCommand(conn)
			if err != nil {
				conn.Close()
				con_clients -= 1
				log.Println("client disconnected : ", conn.RemoteAddr(), " concurrent clients : ", con_clients)
				if err == io.EOF {
					break
				}
				log.Println("err", err)
			}

			log.Println("command : ", cmd)

			respond(cmd, conn)

		}

	}

}

// func readCommand(conn net.Conn) (*core.RedisCmd, error) {

// 	var buf []byte = make([]byte, 512)
// 	n, err := conn.Read(buf[:])

// 	if err != nil {
// 		return nil, err
// 	}
// 	tokens, err := core.DecodeArrayString(buf[:n])

// 	if err != nil {
// 		return nil, err
// 	}

// 	return &core.RedisCmd{
// 		Cmd:  strings.ToUpper(tokens[0]),
// 		Args: tokens[1:],
// 	}, nil

// }
func readCommand(conn io.ReadWriter) (*core.RedisCmd, error) {

	var buf []byte = make([]byte, 512)
	n, err := conn.Read(buf[:])

	if err != nil {
		return nil, err
	}
	tokens, err := core.DecodeArrayString(buf[:n])

	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil

}

func respond(cmd *core.RedisCmd, conn io.ReadWriter) {

	err := core.EvalAndRespond(cmd, conn)

	if err != nil {
		respondError(err, conn)
	}
}

// func respond(cmd *core.RedisCmd, conn net.Conn) {

// 	err := core.EvalAndRespond(cmd, conn)

// 	if err != nil {
// 		respondError(err, conn)
// 	}
// }

func respondError(err error, conn io.ReadWriter) {
	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

// func respondErrorIO(err error, conn io.ReadWriter) {
// 	conn.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
// }
