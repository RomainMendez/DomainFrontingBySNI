package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

// printSeparator prints a separator line with the step name in the middle
func printSeparator(stepName string) {
	width := 80
	title := fmt.Sprintf(" %s ", stepName)
	titleLen := len(title)
	dashCount := (width - titleLen) / 2
	leftDashes := strings.Repeat("-", dashCount)
	rightDashes := strings.Repeat("-", width-dashCount-titleLen)
	fmt.Printf("\n%s%s%s\n", leftDashes, title, rightDashes)
}

// printSuccess prints a success message in green
func printSuccess(message string) {
	fmt.Printf("%s%s%s\n", colorGreen, message, colorReset)
}

// printError prints an error message in red
func printError(message string) {
	fmt.Printf("%s%s%s\n", colorRed, message, colorReset)
}

func main() {
	// Configuration parameters.
	proxyAddr := "localhost:3128"  // Explicit proxy address.
	targetHost := "172.161.32.154" // Domain used in CONNECT request.
	targetPort := "8443"           // Port on which reverse proxy listens.
	connectTarget := targetHost + ":" + targetPort
	sniHost := "example.com" // SNI to be sent during TLS handshake.

	// Proxy authentication credentials.
	proxyUsername := "your-username"
	proxyPassword := "your-password"
	auth := base64.StdEncoding.EncodeToString([]byte(proxyUsername + ":" + proxyPassword))

	// Connect to the explicit proxy.
	printSeparator("CONNECTING TO PROXY")
	fmt.Printf("Connecting to explicit proxy at %s...\n", proxyAddr)
	conn, err := net.DialTimeout("tcp", proxyAddr, 10*time.Second)
	if err != nil {
		printError(fmt.Sprintf("Failed to connect to proxy: %v", err))
		log.Fatalf("Failed to connect to proxy: %v", err)
	}
	printSuccess("Successfully connected to proxy")
	defer conn.Close()

	// Send the CONNECT request to the proxy with Proxy-Authorization header.
	printSeparator("SENDING CONNECT REQUEST")
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Authorization: Basic %s\r\n\r\n", connectTarget, connectTarget, auth)
	fmt.Printf("Sending CONNECT request:\n%s", connectReq)
	_, err = conn.Write([]byte(connectReq))
	if err != nil {
		printError(fmt.Sprintf("Failed to send CONNECT request: %v", err))
		log.Fatalf("Failed to send CONNECT request: %v", err)
	}
	printSuccess("CONNECT request sent successfully")

	// Read the proxy's response.
	printSeparator("READING PROXY RESPONSE")
	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		printError(fmt.Sprintf("Failed to read response from proxy: %v", err))
		log.Fatalf("Failed to read response from proxy: %v", err)
	}
	if !strings.Contains(response, "200") {
		printError(fmt.Sprintf("Proxy did not return 200 OK, got: %s", response))
		log.Fatalf("Proxy did not return 200 OK, got: %s", response)
	}
	// Read any additional header lines until an empty line.
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			printError(fmt.Sprintf("Error reading proxy headers: %v", err))
			log.Fatalf("Error reading proxy headers: %v", err)
		}
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	printSuccess("Tunnel established")

	// Proceed with TLS handshake.
	printSeparator("TLS HANDSHAKE")
	fmt.Println("Proceeding with TLS handshake...")
	tlsConf := &tls.Config{
		ServerName: sniHost,
	}
	tlsConn := tls.Client(conn, tlsConf)
	if err = tlsConn.Handshake(); err != nil {
		printError(fmt.Sprintf("TLS handshake failed: %v", err))
		log.Fatalf("TLS handshake failed: %v", err)
	}
	printSuccess("TLS handshake completed successfully")

	// Send an HTTP GET request over the TLS connection.
	printSeparator("SENDING HTTP REQUEST")
	httpReq := fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", sniHost)
	_, err = tlsConn.Write([]byte(httpReq))
	if err != nil {
		printError(fmt.Sprintf("Failed to send HTTP GET request: %v", err))
		log.Fatalf("Failed to send HTTP GET request: %v", err)
	}
	printSuccess("HTTP request sent successfully")

	// Read and display the HTTP response.
	printSeparator("RECEIVING HTTP RESPONSE")
	resp, err := io.ReadAll(tlsConn)
	if err != nil {
		printError(fmt.Sprintf("Failed to read HTTP response: %v", err))
		log.Fatalf("Failed to read HTTP response: %v", err)
	}
	printSuccess("HTTP response received successfully")

	printSeparator("HTTP RESPONSE CONTENT")
	fmt.Printf("%s\n", string(resp))
}
