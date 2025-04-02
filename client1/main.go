package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	// Read the initial response from the server
	if scanner.Scan() {
		message := scanner.Text()
		if strings.HasPrefix(message, "Error:") {
			log.Printf("Server response: %s", message)
			return
		}
		log.Printf("Server response: %s", message)
	} else if err := scanner.Err(); err != nil {
		log.Printf("Error reading from server: %v", err)
		return
	}

	// Send the client's name
	fmt.Print("Enter your name: ")
	reader := bufio.NewReader(os.Stdin)
	clientName, _ := reader.ReadString('\n')
	clientName = strings.TrimSpace(clientName)

	_, err = conn.Write([]byte(clientName + "\n"))
	if err != nil {
		log.Fatalf("Error sending name: %v", err)
	}

	// Main loop for sending commands
	for {
		fmt.Print("Enter command: ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)

		if command == "exit" {
			log.Println("Exiting...")
			break
		}

		// Send the command to the server
		_, err := conn.Write([]byte(command + "\n"))
		if err != nil {
			log.Printf("Error sending command: %v", err)
			break
		}

		// Wait for the server's response
		if scanner.Scan() {
			response := scanner.Text()
			fmt.Println("Server response:", response)
		} else if err := scanner.Err(); err != nil {
			log.Printf("Error reading response: %v", err)
			break
		} else {
			log.Println("Server disconnected.")
			break
		}
	}
}
