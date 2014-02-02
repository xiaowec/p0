// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type multiEchoServer struct {
	// TODO: implement this!

	conns          map[string]net.Conn    //map of clients
	messages       map[string]chan string //map of channels
	readClientChan chan string            //Channel to receive message from
	//broadcastChan    chan string
	addClientChan    chan net.Conn //Channel to add new client
	removeClientChan chan net.Conn //Channel to remove client
	countClientChan  chan chan int //Channel to get client number
	closeChan        chan int      //Channel to close connection
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	conns := make(map[string]net.Conn, 1)
	messages := make(map[string]chan string, 1)
	readClientChan := make(chan string, 1)
	//broadcastChan := make(chan string, 1)
	addClientChan := make(chan net.Conn, 1)
	removeClientChan := make(chan net.Conn, 1)
	countClientChan := make(chan chan int, 1)
	closeChan := make(chan int, 1)

	return &multiEchoServer{conns, messages, readClientChan, addClientChan,
		removeClientChan, countClientChan, closeChan}
}

func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	fmt.Println("begin listen!")

	//Listen to the port
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	go handleAccept(ln, mes.conns, mes.messages, mes.readClientChan, mes.addClientChan,
		mes.removeClientChan, mes.countClientChan, mes.closeChan)

	return err
}

func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	mes.closeChan <- 1
}

func (mes *multiEchoServer) Count() int {
	// TODO: implement this!
	countchan := make(chan int, 1)
	mes.countClientChan <- countchan
	return <-countchan
}

// TODO: add additional methods/functions below!

func handleAccept(ln net.Listener, conns map[string]net.Conn, messages map[string]chan string,
	readClientChan chan string, addClientChan chan net.Conn, removeClientChan chan net.Conn,
	countClientChan chan chan int, closeChan chan int) {

	// goroutine: Handle client map
	go handleClientMap(conns, messages, addClientChan, removeClientChan, readClientChan,
		countClientChan, closeChan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			break
		} else {
			// add a new client to the map
			addClientChan <- conn

			// goroutine: Read from clients
			go handleConnRead(conn, removeClientChan, readClientChan)
		}
	}
}

func handleWrite(conn net.Conn, message chan string, removeClientChan chan net.Conn) {
	for {
		msg := <-message

		_, err := conn.Write([]byte(msg))
		if err != nil {
			conn.Close()
			removeClientChan <- conn
		}
	}

}

//handle the client map
func handleClientMap(conns map[string]net.Conn, messages map[string]chan string,
	addClientChan chan net.Conn, removeClientChan chan net.Conn,
	readClientChan chan string, countClientChan chan chan int, closeChan chan int) {
	for {
		select {
		case conn := <-addClientChan:
			// add a client to the map of clients.
			conns[conn.RemoteAddr().String()] = conn

			message := make(chan string, 100)
			messages[conn.RemoteAddr().String()] = message

			// goroutine: write message to each client
			go handleWrite(conn, message, removeClientChan)

		case conn := <-removeClientChan:
			// remove a client to the map of clients.
			key := conn.RemoteAddr().String()
			delete(conns, key)
			delete(messages, key)

		case receiveStr := <-readClientChan:
			// put message to each client channel
			for _, message := range messages {
				if len(message) < 100 {
					message <- receiveStr
				}
			}

		case countchan := <-countClientChan:
			// get the number of clients
			countchan <- len(conns)

		case value := <-closeChan:
			// close connection
			value = value
			for _, conn := range conns {
				conn.Close()
				removeClientChan <- conn
			}
			break
		}
	}
}

// handle read from clients
func handleConnRead(conn net.Conn, removeClientChan chan net.Conn, readClientChan chan string) {

	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			// remove client from current map
			conn.Close()
			removeClientChan <- conn
			break
		}

		// send string to readclient channel
		receiveStr := string(line)
		readClientChan <- receiveStr
	}

}
