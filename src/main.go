package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
	"os/exec"
)

var (
	localport          = 8080
	proxyHost          = ""
	remoteHost         = ""
	proxyAuthorization = ""
	loginScript = "true"
)

func HandleRequest(clientConn net.Conn) {
	counter := 0
retry:
	timeout,_:= time.ParseDuration("1m");
	if proxyConn, err := net.DialTimeout("tcp", proxyHost,timeout); err != nil {
		fmt.Println("err: dial "+ proxyHost)
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
			if strings.Contains(response,"501") {
				proxyConn.Close()
				err:= exec.Command("sh", loginScript).Run()
				if err != nil{
					fmt.Println("error: exec login-script")
					counter = 5
				} else {
					counter++
					if counter < 2{
						goto retry
					}
				}
			}
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
	_loginScript := flag.String("l","true","/usr/bin/login.sh")
	flag.Parse()
	localport = *_localport
	remoteHost = *_remoteHost
	proxyHost = *_proxyHost
	loginScript = *_loginScript

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
