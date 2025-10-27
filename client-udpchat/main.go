package main

import (
	"fmt"
	"os"
	"net"
	"bufio"
	//"strings"
)

var Debug bool


func main() {

	fmt.Println("==== Welcome! UPD chatter! ====")
	fmt.Println(" >> connecting to rendezvous")
	fmt.Println("    please waiting          ")

	// Rendezvous server address
	rdvAddr := net.UDPAddr{
		IP: net.ParseIP("34.172.225.134"),
		Port: 55585,
	}

	Debug = false

	sPort := "50001"
	dPort := "50002"

	laddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + sPort)
	if err != nil { fmt.Printf("(rdv)address parse failed: %v\n", err) }

	rdvConn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("(rdv)binding failed: %v\n", err) }

	peerIP := string(waitForRdvReply(rdvConn, &rdvAddr))
	
	// Printing the bytes here really helped reveal why the 
	// IP didn't parse (I was trying to parse a []byte of size 512,
	//                  So the IP was padded with zeroes.)
	//fmt.Printf("bytes: [%v]\n", []byte(peerIP))

	fmt.Printf(" >> Peer found: [%s]\n", peerIP)
	fmt.Printf(" >> source port: \t50001\n")
	fmt.Printf(" >> dest port: \t50002\n")


	            // After Server Connect

	scanner := bufio.NewScanner(os.Stdin)

	// Punch hole with some message
	fmt.Printf(" >> punching hole\n")

	laddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0:" + sPort)
	if err != nil { fmt.Printf("(punch)address parse failed: %v\n", err) }

	pconn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("(punch)binding failed: %v\n", err) }

	premote, err := net.ResolveUDPAddr("udp4", peerIP + ":" + dPort)
	if err != nil { fmt.Printf("(punch)address parse failed: %v\n", err) }

	//fmt.Printf(" >> (confirm)Remote IP: [%s][%d]\n", premote.IP.String(), premote.Port)

	pconn.WriteToUDP([]byte("punch"), premote)
	pconn.Close()


	go listenToPort(sPort)
	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Ready, set, chat! ---")
	fmt.Println("--- (ctrl-c to quit)  ---")

	fmt.Printf("\n> ")


	// Loop sending msgs from stdin
	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + dPort)
	if err != nil { fmt.Printf("(main)address parse failed: %v\n", err) }

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil { fmt.Printf("(main)binding failed: %v\n", err) }

	remote, err := net.ResolveUDPAddr("udp4", peerIP + ":" + sPort)
	if err != nil { fmt.Printf("(main)address parse failed: %v\n", err) }

	for scanner.Scan() {
		input := scanner.Text()
		if len(input) == 0 { continue }

		// Create UDP socket and bind to local port
		_, err = conn.WriteToUDP([]byte(input), remote)
		if err != nil { fmt.Printf("(main)sending failed: %v\n", err) }

		fmt.Printf("> ")
	}

}

func listenToPort(port string) error {
	laddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + port)
	if err != nil { fmt.Printf("(listener)parse failed: %v\n", err) }

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("(listener)binding failed: %v\n", err) }

	defer conn.Close()

	b := make([]byte, 512)

	for {
		n, addr, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(listener)read error: %v\n", err) }

		if Debug {
			fmt.Printf("(%s)", addr)
		}
		fmt.Printf("%50s <\n", string(b[:n]))
	}
}

func waitForRdvReply(conn *net.UDPConn, rdvAddr *net.UDPAddr) []byte {
	defer conn.Close()
	fmt.Printf(" >> waiting for rendezvous\n")

	b := make([]byte, 512)

	conn.WriteToUDP([]byte("69"), rdvAddr)

	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(rdv-reply)read error: %v\n", err) }

		//fmt.Printf(" >> Rendezvous replied: %s\n", string(b[:n]))

		if len(b) > 1 {
			return b[:n]
		}
	}
}
