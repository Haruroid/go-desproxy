package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

var (
	localport          = 8080
	proxyHost          = ""
	remoteHost         = ""
	proxyAuthorization = ""
)

func HandleRequest(clientConn net.Conn) {
	if proxyConn, err := net.Dial("tcp", proxyHost); err != nil {
		log.Fatal(err)
	} else {
		var proxyauth = ""
		if !strings.Contains(remoteHost,"maizuru") {
			proxyauth = fmt.Sprintf("Proxy-Authorization: %s",proxyAuthorization) + "\r\n"
		}
		var request = fmt.Sprintf("CONNECT %s HTTP/1.0\r\n%s\r\n",remoteHost,proxyauth)
		proxyConn.Write([]byte(request))
		
		scanner := bufio.NewScanner(proxyConn)
		scanner.Scan()
		var response = scanner.Text()
		if !strings.Contains(response,"200") {
			fmt.Println("err: "+response)
			proxyConn.Close()
			clientConn.Close()
		}
		go func() {
			io.Copy(clientConn, proxyConn)
			proxyConn.Close()
		}()
		go func() {
			io.Copy(proxyConn, clientConn)
			clientConn.Close()
		}()
	}
}

func main() {
	_proxyUser := flag.String("u", "", "username:password")
	_localport := flag.Int("p", 8080, "local port")
	_remoteHost := flag.String("r", "", "remote host:port")
	_proxyHost := flag.String("x", "10.1.16.8:8080", "Proxy:port")
	flag.Parse()
	localport = *_localport
	remoteHost = *_remoteHost
	proxyHost = *_proxyHost

	proxyAuthorization = "Basic " + base64.StdEncoding.EncodeToString([]byte(*_proxyUser))
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", localport))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Println("Listening on localhost:")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		go HandleRequest(conn)
	}
}
