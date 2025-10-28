package main

import (
	"fmt"
	"os"
	"net"
	"bufio"
	"golang.org/x/term"
)

var Debug bool

type TermInfo struct {
	fd int
	Cols int
	Rows int
}


func main() {
	enableVirtualTerminalProcessing()
	version := "v0.3"
	termInfo := TermInfo{}

	fmt.Println("==== Welcome! UPD chatter! ====")
	fmt.Printf("                           %s   \n", version)
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
	if err != nil { fmt.Printf("(rdv)binding failed: %v\n", err); return }

	peerIP := string(waitForRdvReply(rdvConn, &rdvAddr))
	
	// Printing the bytes here really helped reveal why the 
	// IP didn't parse (I was trying to parse a []byte of size 512,
	//                  So the IP was padded with zeroes.)
	//fmt.Printf("bytes: [%v]\n", []byte(peerIP))

	fmt.Printf(" >> Peer found: \t[%s]\n", peerIP)
	fmt.Printf("    source port:\t50001\n")
	fmt.Printf("    dest port:  \t50002\n")


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
	fmt.Println("---    /q to quit     ---")

	initTerminal(&termInfo)

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

		n, err := conn.WriteToUDP([]byte(input), remote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }

		if input == "/q" {
			conn.WriteToUDP([]byte(" >> Peer disconnected"), remote)
			err := conn.Close()
			if err != nil { fmt.Printf("(main)could not close connection\nports 50001 and 50002 possibly remain bound: %v\n", err) }
			fmt.Printf("\n>> Good bye!\n")
			restoreTerminal()
			return
		}

		
		fmt.Printf("\033[1F\n> %s", input)// Print user message above input field
		fmt.Printf("\033[%d;0H", termInfo.Rows)  // Move to bottom left
		fmt.Printf("\033[0K")           // Clear to EOL
		fmt.Printf("> ")                // Print Cursor
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
		fmt.Printf("\0337")
		fmt.Printf("\033[1A")
		fmt.Printf("\n%50s <", string(b[:n]))
		fmt.Printf("\0338")
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

func initTerminal(termInfo *TermInfo) {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil { fmt.Printf("(init)failed to get terminal size: %v\n", err) }

	//fmt.Printf("Cols: %d\n", cols)
	//fmt.Printf("Rows: %d\n", rows)

	termInfo.Cols = cols
	termInfo.Rows = rows


	// Save Cursor position
	fmt.Printf("\0337")

	// Set scrollable region:
	fmt.Printf("\033[0;%dr", rows - 1)

	// Move the cursor to bottom
	fmt.Printf("\033[%d;0H", rows)
}

func restoreTerminal() {
	//fmt.Printf("\0338")
}
