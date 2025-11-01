package main

import (
	"fmt"
	"os"
	"net"
	"bufio"
	"strings"
	"golang.org/x/term"
)

var Debug bool

type TermInfo struct {
	fd int
	Cols int
	Rows int
	oldState *term.State
}


func main() {
	enableVirtualTerminalProcessing()
	version := "v0.4"
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

	 // Bound Source Port (Listen)
	rdvConn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("(rdv)binding failed: %v\n", err); return }

	peerPubIP, _ := waitForRdvReply(rdvConn, rdvAddr)
	
	// Printing the bytes here really helped reveal why the 
	// IP didn't parse (I was trying to parse a []byte of size 512,
	//                  So the IP was padded with zeroes.)
	// fmt.Printf("bytes: [%v]\n", []byte(peerIP))

	fmt.Printf(" >> Peer found: \t[%s]\n", peerPubIP)
	fmt.Printf("    source port:\t50001\n")
	fmt.Printf("    dest port:  \t50002\n")


	 // After Server Connect
	scanner := bufio.NewScanner(os.Stdin)

     // Punch hole
	fmt.Printf(" >> punching hole\n")

	laddr, err = net.ResolveUDPAddr("udp4", "0.0.0.0:" + sPort)
	if err != nil { fmt.Printf("(punch)address parse failed: %v\n", err) }

	pconn, err := net.ListenUDP("udp4", laddr)
	if err != nil { fmt.Printf("(punch)binding failed: %v\n", err) }

	premote, err := net.ResolveUDPAddr("udp4", peerPubIP + ":" + dPort)
	if err != nil { fmt.Printf("(punch)address parse failed: %v\n", err) }


	 // Send packet - From S -> D
	 // Listening to *all* on port S.
	pconn.WriteToUDP([]byte("punch"), premote)
	pconn.Close() // Why do we close here?


	go listenToPort(sPort)  // Pretty sure we just reopen that same port here.
	fmt.Printf(" >> Listening...\n\n")
	fmt.Println("--- Ready, set, chat! ---")
	fmt.Println("---    /q to quit     ---")

	initTerminal(&termInfo)

	fmt.Printf("\n> ")


	addr, err := net.ResolveUDPAddr("udp4", "0.0.0.0:" + dPort)
	if err != nil { fmt.Printf("(main)address parse failed: %v\n", err) }

	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		fmt.Printf("(main)binding failed: %v\n", err)
		return
	}

	remote, err := net.ResolveUDPAddr("udp4", peerPubIP + ":" + sPort)
	if err != nil { fmt.Printf("(main)address parse failed: %v\n", err) }

	// Loop sending msgs from stdin
	for scanner.Scan() {
		input := scanner.Text()
		if len(input) == 0 { continue }

		n, err := conn.WriteToUDP([]byte(input), remote)
		if err != nil { fmt.Printf("(main)sending [%d bytes] failed: %v\n", n, err) }

		if input == "/q" {
			conn.WriteToUDP([]byte(" >> Peer disconnected"), remote)
			err := conn.Close()
			// TODO: add real port info
			if err != nil {
				fmt.Printf("(main)could not close connection\nports %s and %s possibly remain bound: %v\n",
					err, "TODO", "TODO")}
			fmt.Printf("\n>> Good bye!\n")
			restoreTerminal(&termInfo)
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
			fmt.Printf("\0337")
			fmt.Printf("\033[1A")
			fmt.Printf("\n%s:%d", addr.IP, addr.Port)
			fmt.Printf("\0338")
		}

		fmt.Printf("\0337")
		fmt.Printf("\033[1A")
		fmt.Printf("\n%50s <", string(b[:n]))
		fmt.Printf("\0338")
	}
}


func waitForRdvReply(conn *net.UDPConn, rdvAddr net.UDPAddr) (string, string) {
	peerPublicEndpoint := ""
	peerPrivateEndpoint := ""
	defer conn.Close()
	fmt.Printf(" >> waiting for rendezvous\n")

	b := make([]byte, 65507)

	// Here, instead of sending funny number, we will send something *useful*
    // We need to send the endpoint we believe 
	// we are using to communicate with the server.
	privEndpoint := conn.LocalAddr().String()
	conn.WriteToUDP([]byte(privEndpoint), &rdvAddr)

	for {
		n, _, err := conn.ReadFromUDP(b)
		if err != nil { fmt.Printf("(rdv-reply)read error: %v\n", err) }

		//fmt.Printf(" >> Rendezvous replied: %s\n", string(b[:n]))

		if len(b) > 1 {

			if peerPublicEndpoint != "" && peerPrivateEndpoint != "" {
				return peerPublicEndpoint, peerPrivateEndpoint
			}

			data, found := strings.CutPrefix(string(b[:n]), "peerPublicEndpoint:")
			if found {
				addr, err := net.ResolveUDPAddr(conn.RemoteAddr().Network(), data)	

				if err != nil {
				 	fmt.Printf("(rdv-reply)read error: %v\n", err)
					return "not", "good"
				}

				peerPublicEndpoint = addr.String()
				continue
			}

			data, found = strings.CutPrefix(string(b[:n]), "peerPrivateEndpoint:") 
			if found {
				addr, err := net.ResolveUDPAddr( conn.RemoteAddr().Network(), data)	
				if err != nil {
					fmt.Printf("(rdv-reply)read error: %v\n", err)
					return "not", "good"
				}

				peerPrivateEndpoint = addr.String()
			}

		}

	}
}

func initTerminal(termInfo *TermInfo) {
	cols, rows, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil { fmt.Printf("(init)failed to get terminal size: %v\n", err) }

	oldState, err := term.GetState(int(os.Stdout.Fd()))
	if err != nil { fmt.Printf("(init)failed to get terminal state: %v\n", err) }

	termInfo.oldState = oldState
	termInfo.fd = int(os.Stdout.Fd())

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

func restoreTerminal(t *TermInfo) {
	term.Restore(t.fd, t.oldState)
}
