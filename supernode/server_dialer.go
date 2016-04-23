package main

import (
	"bufio"
	"dsproject/util"
	"fmt"
	"io"
	"math"
	"net"
	"strconv"
	"strings"
	"time"
)

func dialServer() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", util.ServerAddr)
	util.CheckError(err)

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	util.CheckError(err)

	_, err = conn.Write([]byte("SUPERNODE REGISTER " + port + "\r\n"))
	util.CheckError(err)

	// send Heartbeat
	go func() {
		for {
			writer := bufio.NewWriter(conn)
			writer.WriteString("SUPERNODE HEARTBEAT " + port + "\n")
			writer.Flush()
			time.Sleep(1000 * time.Millisecond)
		}
	}()

	reader := bufio.NewReader(conn)
	// Read handler
	for {
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		util.CheckError(err)

		fmt.Println("[Server Message]:" + message)
		processCommand(message)
	}
}

// process command from Server
func processCommand(cmd string) {
	args := strings.Split(strings.Trim(cmd, "\r\n"), " ")

	//Compute distance to the customer
	if args[0] == "PEERADDR" {
		peerAddr := args[1]
		go dialPeer(peerAddr)
	}

	//Compute distance to the customer
	if args[0] == "PICKUP" {
		//Pickup the customer
		//source := util.ParseFloatCoordinates(args[1], args[2])
		//dest := util.ParseFloatCoordinates(args[3], args[4])
		source := args[1] + " " + args[2]
		dest := args[3] + " " + args[4]
		// if source == nil || dest == nil {
		// 	fmt.Println("Error: incorrect PICKUP format:" + cmd)
		// 	return
		// }

		fmt.Println("In processCommand: " + strconv.Itoa(COUNTCAR))
		request := util.Request{math.MaxFloat64, nil, "", COUNTCAR, source, dest}

		fmt.Println("args[5] =  " + args[5])
		REQMAP[args[5]] = request
		fmt.Println("In MAP: " + strconv.Itoa(REQMAP[args[5]].Count))

		fmt.Println("ID:" + args[5] + " counter" + strconv.Itoa(COUNTCAR))

		var zero []byte
		for _, client := range clients {
			if client.Type == "NODE" {
				continue
			}
			fmt.Println("client " + client.Type + " " + client.Name)
			conn := client.Conn
			fmt.Println(conn.RemoteAddr().String())
			reader := bufio.NewReader(conn)
			_, err := reader.Read(zero)
			if err != nil {
				continue
			}

			fmt.Println("[COMPUTE] send to CarNode:" + client.Conn.RemoteAddr().String())
			writer := bufio.NewWriter(conn)
			writer.WriteString("COMPUTE " + args[1] + " " + args[2] + " " + args[5] + "\n")
			writer.Flush()
		}
	}
}
