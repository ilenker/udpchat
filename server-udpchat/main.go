package main

import (
	"fmt"
	"net"
)

var Debug bool

func main() {
	
	fmt.Println("===    *NIMBUSSY*     ===")
	fmt.Println("=== Rendezvous Server ===")
	fmt.Println("===      v0.01        ===")
	fmt.Println()
	fmt.Println(" listening on port 55585")
	fmt.Println()
	port := "55585"

	clients := make([]net.UDPAddr, 0, 2)

	laddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + port)
	if err != nil { fmt.Printf("listener parse failed: %v\n", err) }

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("listener binding failed: %v\n", err) }

	defer conn.Close()

	b := make([]byte, 512)

	for {
		_, addr, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("listener read error: %v\n", err) }
		
		fmt.Println("o---------------------------o")
		fmt.Printf("Client [%d] connected:\n", len(clients))
		fmt.Printf("\tIP:\t [%s]\n", addr.IP.String())
		fmt.Printf("\tPort:\t [%d]\n", addr.Port)
		fmt.Println("o---------------------------o")
		fmt.Println()
		
		clients = append(clients, *addr)
		if len(clients) == 2 {
			fmt.Println("o---------------------------o")
			fmt.Println("       !Lobby Filled!")
			fmt.Println("  Exchanging IPs and ports.")
			fmt.Println("o---------------------------o")
			break
		}
	}
	c0 := fmt.Sprintf("%s:%d", clients[0].IP.String(), clients[0].Port)
	c1 := fmt.Sprintf("%s:%d", clients[1].IP.String(), clients[1].Port)
	conn.WriteToUDP([]byte(c1), &clients[0])	
	conn.WriteToUDP([]byte(c0), &clients[1])	
}

