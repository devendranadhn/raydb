package server

import (
	"log"
	"net"
	"ray/config"
	"ray/core"
	"syscall"
	"time"
)

var con_clients int = 0
var cronFrequency time.Duration = time.Second
var lastCronExecTime time.Time = time.Now()

func RunASyncTCPServer() error {
	log.Println("starting the asynchronous TCP server on : ", config.Host, config.Port)

	maxclients := 20000

	// create epoll event objects to hold epoll events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, maxclients)

	// create a socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(serverFD)

	// set the socket to operate in non-blocking mode
	if err = syscall.SetNonblock(serverFD, true); err != nil {
		return err
	}

	// bind the ip and port
	ip4 := net.ParseIP(config.Host)
	if err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); err != nil {
		return err
	}

	// start listenting
	if err = syscall.Listen(serverFD, maxclients); err != nil {
		return err
	}

	// async IO starts here

	//creating epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Fatal(err)
	}
	defer syscall.Close(epollFD)

	// specify the events that we want to get the hint about
	// and set the socket on which
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// listen to read events on the server itself
	if err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent); err != nil {
		return err
	}

	for {

		// since it's single threaded, try deleting the random sample of keys for each configured freq
		if time.Now().After(lastCronExecTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecTime = time.Now()
		}

		// see if any fd is ready
		nevents, err := syscall.EpollWait(epollFD, events[:], -1)
		if err != nil {
			continue
		}

		for i := 0; i < nevents; i++ {

			// if the socket server itself is ready for an IO
			if int(events[i].Fd) == serverFD {

				// accept the incoming tcp connection from a client
				fd, _, err := syscall.Accept(serverFD)

				if err != nil {
					log.Print("err", err)
					continue
				}

				// increase the no of con_clients count
				con_clients++
				log.Println("client connected, no of concurrent clients : ", con_clients)
				syscall.SetNonblock(serverFD, true)

				// add this new connection to be monitored
				var serverClientSocket syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &serverClientSocket); err != nil {
					log.Fatal(err)
				}
			} else {
				comm := core.FDComm{Fd: int(events[i].Fd)}
				cmd, err := readCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					con_clients--
					continue
				}
				respond(cmd, comm)
			}

		}
	}

}

// func readCommandIO(conn io.ReadWriter) (*core.RedisCmd, error) {

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
