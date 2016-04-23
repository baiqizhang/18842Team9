package main

import (
	"bufio"
	"container/list"
	"dsproject/util"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var virtualCar util.VirtualCar

func main() {
	// Check arguments
	args := os.Args[1:]
	if len(args) < 3 {
		fmt.Println("Usage: carnode XCOORD YCOORD NAME [-v]")
		os.Exit(0)
	}
	for _, arg := range args {
		if arg == "-v" {
			util.Verbose = 1
		}
	}

	// set car current position
	ptrPoint := util.ParseFloatCoordinates(args[0], args[1])
	if ptrPoint == nil {
		fmt.Println("Error: incorrect XCOORD format")
		os.Exit(0)
	}
	virtualCar.Location = *ptrPoint
	virtualCar.Idle = true
	virtualCar.Name = args[2]

	// Get supernode addresses
	supernodes := getSupernodesAddr()
	for e := supernodes.Front(); e != nil; e = e.Next() {
		fmt.Println("[SuperNode Addr]" + e.Value.(string))
	}
	if supernodes.Len() == 0 {
		fmt.Println("No supernode available, exit")
		os.Exit(0)
	}

	//connect to the first supernode
	dialSuperNode(supernodes.Front().Value.(string))
}

func getSupernodesAddr() list.List {
	var ipList list.List
	conn, err := net.Dial("tcp", util.ServerAddr)
	for err != nil {
		fmt.Println("Unable to connect to the front-end server")
		conn, err = net.Dial("tcp", util.ServerAddr)
		time.Sleep(200 * time.Millisecond)
	}

	conn.Write([]byte("NODE REGISTER CAR1\n"))

	// build a table of supernodes' IP
	connbuf := bufio.NewReader(conn)
	for {
		ip, err := connbuf.ReadString('\n')
		if err != nil || ip == "OK\n" {
			break
		}
		ipList.PushBack(ip[0 : len(ip)-1])
	}
	conn.Close()

	return ipList
}

func dialSuperNode(supernode string) {
	conn, err := net.Dial("tcp", supernode)

	for err != nil {
		conn, err = net.Dial("tcp", supernode)
	}

	conn.Write([]byte("CARNODE\n"))

	connbuf := bufio.NewReader(conn)
	for {
		cmd, _ := connbuf.ReadString('\n')
		processCommand(cmd, conn)
	}
}

//processCommand Process commands that's received by car node
func processCommand(cmd string, conn net.Conn) {
	args := strings.Split(strings.Trim(cmd, "\r\n"), " ")

	//Compute distance to the customer
	if args[0] == "COMPUTE" {
		var distance float64

		//if not idle, return -1
		if virtualCar.Idle {
			point := util.ParseFloatCoordinates(args[1], args[2])
			distance = point.DistanceTo(virtualCar.Location)
		} else {
			distance = math.MaxFloat64
		}
		writer := bufio.NewWriter(conn)
		fmt.Println(args[3])
		writer.WriteString("COMPUTERESULT " + virtualCar.Name + " " + strconv.FormatFloat(distance, 'f', 4, 64) + " " + args[1] + " " + args[2] + " " + args[3] + "\n")
		writer.Flush()
	} else if args[0] == "PICKUP" {
		//Pickup the customer
		source := util.ParseFloatCoordinates(args[1], args[2])
		dest := util.ParseFloatCoordinates(args[3], args[4])
		if source == nil || dest == nil {
			fmt.Println("Error: incorrect PICKUP format")
			os.Exit(0)
		}

		//Start simulation
		go util.DriveCustomer(&virtualCar, source, dest)
	}
}
