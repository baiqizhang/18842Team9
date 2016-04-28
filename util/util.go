package util

import (
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"time"
)

// Verbose verbose mode
var Verbose = 0

//PickupToken a token structure that will be passed among supernodes for visualization and pickup
type PickupToken struct {
	ReqID  int
	Origin string
	Src    Point
	Dest   Point
	Length int
	Points []Point
	Addrs  []string
}

/* FailureToken which gets passed around during ring repair */
type FailureToken struct {
	FailAddr      string
	InitiatedNode string
}

/*
	Communication
*/

//ServerAddr address of the central server
const ServerAddr = "127.0.0.1:7070"

//Client a struct that models a TCP Client by connection and name
type Client struct {
	Conn net.Conn
	Name string
	Type string
}

/*
	Struct Point
*/

//Point p(x,y) struct
type Point struct {
	X float64
	Y float64
}

//DistanceTo compute the distance from V1 to V2
func (v1 *Point) DistanceTo(v2 Point) float64 {
	return math.Sqrt((v1.X-v2.X)*(v1.X-v2.X) + (v1.Y-v2.Y)*(v1.Y-v2.Y))
}

//ParseFloatCoordinates construct a Point using two strings
func ParseFloatCoordinates(strx string, stry string) *Point {
	x, err := strconv.ParseFloat(strx, 64)
	if err != nil {
		fmt.Println("x coordinate is not a valid float value")
		return nil
	}

	y, err := strconv.ParseFloat(stry, 64)
	if err != nil {
		fmt.Println("y coordinate is not a valid float value")
		return nil
	}
	result := Point{X: x, Y: y}
	return &result
}

/*
	Struct VirtualCar
*/

//VirtualCar simulated car instance
type VirtualCar struct {
	Name     string
	Location Point
	Idle     bool
}

//struct for request
type Request struct {
	FinalResult float64
	FinalConn   net.Conn
	Carname     string
	Count       int
	Source      string
	Dest        string
}

//DriveCustomer simulate a ride from currentLoc to customerLoc then to destLoc
func DriveCustomer(car *VirtualCar, customerLoc *Point, dest *Point) {
	car.Idle = false
	fmt.Println("[Simulation] Car on the way")

	dt := 100.0

	diffx := (customerLoc.X - car.Location.X) / dt
	diffy := (customerLoc.Y - car.Location.Y) / dt
	// simulate picking up customer
	for i := 0; i < int(dt); i++ {
		time.Sleep(100 * time.Millisecond)
		car.Location.X = car.Location.X + diffx
		car.Location.Y = car.Location.Y + diffy
	}

	// update current location
	car.Location = *customerLoc
	fmt.Println("[Simulation] Customer picked up")

	diffx = (dest.X - car.Location.X) / dt
	diffy = (dest.Y - car.Location.Y) / dt
	// simulate picking up customer
	for i := 0; i < int(dt); i++ {
		time.Sleep(100 * time.Millisecond)
		car.Location.X = car.Location.X + diffx
		car.Location.Y = car.Location.Y + diffy
	}

	fmt.Println("[Simulation] Drop customer")

	car.Location = *dest
	car.Idle = true
}

/*
	Common Utils
*/

//CheckError just check and print error
func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s\n", err.Error())
		os.Exit(1)
	}
}
