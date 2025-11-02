package main

import (
	"fmt"
	"net"
)

type ClientEndpoints struct{
	pubEndpoint net.UDPAddr
	privEndpoint net.UDPAddr
}

func main() {
	
	fmt.Println("===    *NIMBUSSY*     ===")
	fmt.Println("=== Rendezvous Server ===")
	fmt.Println("===      v0.02        ===")
	fmt.Println()
	fmt.Println(" listening on port 55585")
	fmt.Println()
	port := "55585"

	clients := make([]ClientEndpoints, 0, 2)

	laddr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + port)
	if err != nil {
		fmt.Printf("listener parse failed: %v\n", err)
		return
	}

	conn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		fmt.Printf("listener binding failed: %v\n", err)
		return
	}

	defer conn.Close()

	b := make([]byte, 65507)

	for {
		n, pubEndpoint, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("listener read error: %v\n", err) }

		privEndpoint, err := net.ResolveUDPAddr("udp4", string(b[:n]))
		if err != nil {
			fmt.Printf("privEndpoint parse failed: %v\n", err)
			return
		}
		
		fmt.Println("o---------------------------o")
		fmt.Printf("Client [%d] connected:\n", len(clients))
		fmt.Printf("\tPublic  IP  : [%s]\n", pubEndpoint.IP.String())
		fmt.Printf("\tPublic  Port: [%d]\n", pubEndpoint.Port)
		fmt.Println("             ~~~             ")
		fmt.Printf("\tPrivate IP  : [%s]\n", privEndpoint.IP.String())
		fmt.Printf("\tPrivate Port: [%d]\n", privEndpoint.Port)
		fmt.Println("o---------------------------o")
		fmt.Println()
		
		clients = append(clients, ClientEndpoints{
			pubEndpoint: *pubEndpoint,
			privEndpoint: *privEndpoint,
		})

		if len(clients) == 2 {
			fmt.Println("o---------------------------o")
			fmt.Println("       !Lobby Filled!")
			fmt.Println("  Exchanging IPs and ports.")
			fmt.Println("o---------------------------o")
			break
		}
	}
	c0pub  := fmt.Sprintf("peerPublicEndpoint:%s:%d",
		clients[0].pubEndpoint.IP.String(),
		clients[0].pubEndpoint.Port,)
	c0priv := fmt.Sprintf("peerPrivateEndpoint:%s:%d",
		clients[0].privEndpoint.IP.String(),
		clients[0].privEndpoint.Port)
	c1pub  := fmt.Sprintf("peerPublicEndpoint:%s:%d",
		clients[1].pubEndpoint.IP.String(),
		clients[1].pubEndpoint.Port)
	c1priv := fmt.Sprintf("peerPrivateEndpoint:%s:%d",
		clients[1].privEndpoint.IP.String(),
		clients[1].privEndpoint.Port)

	// hacky way to detect local peers and swap local IPs
	if clients[0].pubEndpoint.IP.String() == clients[1].pubEndpoint.IP.String() {
		c0pub  = fmt.Sprintf("peerPrivateEndpoint:%s:%d",
			clients[0].pubEndpoint.IP.String(),
			clients[0].pubEndpoint.Port,)
		c0priv = fmt.Sprintf("peerPublicEndpoint:%s:%d",
			clients[0].privEndpoint.IP.String(),
			clients[0].privEndpoint.Port)
		c1pub  = fmt.Sprintf("peerPrivateEndpoint:%s:%d",
			clients[1].pubEndpoint.IP.String(),
			clients[1].pubEndpoint.Port)
		c1priv = fmt.Sprintf("peerPublicEndpoint:%s:%d",
			clients[1].privEndpoint.IP.String(),
			clients[1].privEndpoint.Port)
	} 

	conn.WriteToUDP([]byte(c1pub), &clients[0].pubEndpoint)	
	conn.WriteToUDP([]byte(c1priv), &clients[0].pubEndpoint)	
	conn.WriteToUDP([]byte(c0pub), &clients[1].pubEndpoint)	
	conn.WriteToUDP([]byte(c0priv), &clients[1].pubEndpoint)	
}

