package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// Config holds the desired server configuration
type Config struct {
	MaxClients int `json:"max_clients"`
}

// Load configuration from a JSON file
func loadConfig(filename string) (Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

// Handle a client connection
func handleConnection(conn net.Conn, config Config, clientCount *int, mutex *sync.Mutex) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	// Check if the server is full
	mutex.Lock()
	if *clientCount >= config.MaxClients {
		conn.Write([]byte("Error: Server is full. Try again later.\n"))
		mutex.Unlock()
		return
	}

	// At this point we accepted the client
	*clientCount++
	mutex.Unlock()

	log.Printf("Client connected: %s", clientAddr)
	conn.Write([]byte("Welcome to the server! Please provide your name.\n"))

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		log.Printf("Client disconnected before sending a name: %s", clientAddr)
		mutex.Lock()
		*clientCount--
		mutex.Unlock()
		return
	}
	clientName := scanner.Text()
	log.Printf("Client %s connected with name: %s", clientAddr, clientName)

	for scanner.Scan() {
		request := scanner.Text()
		log.Printf("Client %s (%s) sent request: %s", clientAddr, clientName, request)

		// Process the request
		response := processRequest(request)

		log.Printf("Server processed request from %s (%s): %s", clientAddr, clientName, response)
		_, err := conn.Write([]byte(response + "\n"))
		if err != nil {
			log.Printf("Error sending response to client %s (%s): %v", clientAddr, clientName, err)
			break
		}
	}

	log.Printf("Client disconnected: %s (%s)", clientAddr, clientName)
	mutex.Lock()
	*clientCount--
	mutex.Unlock()
}

// Proccess incoming requests
func processRequest(request string) string {
	parts := strings.Fields(request)
	if len(parts) < 1 {
		return "Error: invalid command"
	}

	command := parts[0]
	switch command {
	case "createWords":
		return createWords(parts)
	case "countPerfectSquares":
		return countPerfectSquares(parts)
	case "convert2to10":
		return convertBinaryToDecimal(parts)
	case "decodeText":
		return decodeText(parts)
	case "checkVowels":
		return checkVowels(parts)
	case "sum":
		return sum(parts)
	case "validPass":
		return validPasswords(parts)
	default:
		return "Error: unknown command"
	}
}

// Server functionalities -------------- START

// take the character at the index the word is in the array
func createWords(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage createWords <string>"
	}

	parts = parts[1:]

	maxLength := len(parts[0])
	for _, word := range parts {
		if len(word) > maxLength {
			maxLength = len(word)
		}
	}

	var result string

	for i := range maxLength {
		var word string

		for _, wordStr := range parts {
			if i < len(wordStr) {
				word += string(wordStr[i])
			} else {
				word += " "
			}
		}

		result += word + " "
	}

	return result
}

func isPerfectSquare(n int) bool {
	sqrt := int(math.Sqrt(float64(n)))
	return sqrt*sqrt == n
}

// find numbers in string and look for perfect square numbers
func countPerfectSquares(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage countPerfectSquares <string>"
	}

	count := 0
	numberRegex := regexp.MustCompile(`\d+`)

	for _, s := range parts[1:] {
		matches := numberRegex.FindAllString(s, -1)
		for _, match := range matches {
			num, err := strconv.Atoi(match)
			if err == nil && isPerfectSquare(num) {
				count++
			}
		}
	}

	return fmt.Sprintf("Perfect squares count: %d", count)
}

func isBinary(s string) bool {
	for _, char := range s {
		if char != '0' && char != '1' {
			return false
		}
	}
	return true
}

// parse for a binary sequence and convert it in decimal
func convertBinaryToDecimal(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage convertBinaryToDecimal <string>"
	}

	var result []string

	inputStrings := strings.Split(strings.Join(parts[1:], " "), ",")
	for _, s := range inputStrings {
		s = strings.TrimSpace(s)
		if isBinary(s) {
			decimal, err := strconv.ParseInt(s, 2, 64)
			if err == nil {
				result = append(result, fmt.Sprintf("%d", decimal))
			}
		}
	}

	if len(result) == 0 {
		return "No valid binary strings found."
	}

	return fmt.Sprintf("Converted values: %s", strings.Join(result, ", "))
}

// decode words formatted (number)text into text * number (ex. 3aba => abaabaaba)
func decodeText(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage decodeText <encoded_text>"
	}

	encodedText := parts[1]
	var decodedText strings.Builder

	i := 0
	for i < len(encodedText) {
		start := i
		for i < len(encodedText) && unicode.IsDigit(rune(encodedText[i])) {
			i++
		}

		if start == i || i >= len(encodedText) {
			return "Error: invalid encoded text"
		}

		count, err := strconv.Atoi(encodedText[start:i])
		if err != nil || count < 1 {
			return "Error: invalid repetition count"
		}

		char := encodedText[i]
		decodedText.WriteString(strings.Repeat(string(char), count))

		i++
	}

	return decodedText.String()
}

func isVowel(c rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, c)
}

// check that the string has even number of vowels on even indexes
func checkVowels(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage checkVowels <string>"
	}

	input := strings.Join(parts[1:], " ")

	words := strings.Split(input, ",")
	var validWords []string

	for _, word := range words {
		word = strings.TrimSpace(word)
		vowelCount := 0

		for i, c := range word {
			if i%2 == 0 && isVowel(c) {
				vowelCount++
			}
		}

		if vowelCount > 0 && vowelCount%2 == 0 {
			validWords = append(validWords, word)
		}
	}

	if len(validWords) == 0 {
		return "Nu sunt cuvinte valide"
	}

	return strings.Join(validWords, " ")
}

// Take an array of strings, parse to ints, double the first digit and calculate the sum
func sum(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage sum <num1>, <num2>, ... 	"
	}

	sum := 0

	for i := 1; i < len(parts)-1; i++ {
		parts[i] = strings.TrimSpace(parts[i])
		if len(parts[i]) == 0 {
			continue
		}
		fmt.Println(parts[i])
		firstDigit, err := strconv.Atoi(string(parts[i][0]))
		if err != nil {
			return "Error: invalid number in input"
		}
		fmt.Println(firstDigit)
		newNumStr := strconv.Itoa(firstDigit) + parts[i]
		newNum, _ := strconv.Atoi(newNumStr)

		sum += newNum
	}

	return fmt.Sprintf("Suma este: %d", sum)
}

// Check if a password follows certain rules
func validPasswords(parts []string) string {
	if len(parts) < 2 {
		return "Error: usage validPass <string>"
	}

	var re = regexp.MustCompile(`^[a-zA-Z0-9!@#$%^&*()_+=-]+$`)

	var validPasswords []string

	for _, password := range parts {
		if re.MatchString(password) {
			hasLower := false
			hasUpper := false
			hasDigit := false
			hasSpecial := false

			for _, char := range password {
				if char >= 'a' && char <= 'z' {
					hasLower = true
				}
				if char >= 'A' && char <= 'Z' {
					hasUpper = true
				}
				if char >= '0' && char <= '9' {
					hasDigit = true
				}
				if strings.ContainsRune("!@#$%^&*()_+=-", char) {
					hasSpecial = true
				}
			}

			if hasLower && hasUpper && hasDigit && hasSpecial {
				validPasswords = append(validPasswords, password)
			}
		}
	}
	return strings.Join(validPasswords, ", ")
}

// Server functionalities -------------- END

func main() {
	config, err := loadConfig("./server_config.json")
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer listener.Close()

	log.Printf("Server started on port 8080 with max clients: %d", config.MaxClients)

	clientCount := 0
	var mutex sync.Mutex

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn, config, &clientCount, &mutex)
	}
}
