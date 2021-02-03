package main

import (
	//"bytes"
	"fmt"
	"flag"
	//"io/ioutil"
	//"log"
	"net"
	"time"
	"math/rand"
	"github.com/gorilla/websocket"
	//"net/http"
	//"net/http/httptrace"
	"strconv"
)

var laddr string
var remoteserverpath string = "client"
var port, remoteserver string

func main() {
	//get client ipv6 addr
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())
		if ip.To4() == nil && ip.IsGlobalUnicast() {
			laddr = ip.String()
			fmt.Println(laddr)
			break
		}
	}
	//listen client port
	flag.StringVar(&port, "port", "1022", "default port for ssh")
	flag.StringVar(&remoteserver, "remoteserver", "ws://127.0.0.1:8080/", "default remote server")
	flag.Parse()
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println(err)
	}
	//loop for accept local connetion
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleconn(conn)
	}
}

func handleconn(conn net.Conn) {
	//1.wss to remoteserver
	wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+remoteserverpath, nil)
	if err != nil {
		fmt.Println(err)
	}
	//2.send client ipv6:port to remortserver
	rand.Seed(time.Now().UnixNano())
	lport := rand.Intn(55536) + 10000
	lportstr := strconv.Itoa(lport)
	var sendbuf string = "[" + laddr +"]:" + lportstr
	fmt.Println(sendbuf)
	err = wsconn.WriteMessage(websocket.BinaryMessage, []byte(sendbuf))
	if err != nil {
		fmt.Println("send client ipv6:port to remoteserver err : ",err)
		return
	}
	//3.get localserver ipv6 from remoteserver
	_, recvbuf, err := wsconn.ReadMessage()
	if err != nil {
		fmt.Println("get localserver ipv6 from remoteserver err : ",err)
		return
	}
	raddr := "[" + string(recvbuf) + "]:" + lportstr
	fmt.Println(raddr)
	//4.dial to localserver
	laddrtcp, err := net.ResolveTCPAddr("tcp6", sendbuf)
	raddrtcp, err := net.ResolveTCPAddr("tcp6", raddr)
	lsconn, err := net.DialTCP("tcp6", laddrtcp, raddrtcp)
	//5.send cmd to localserver
	cmdport, _ := strconv.Atoi(port)
	cmdport = cmdport - 1000
	cmd := strconv.Itoa(cmdport)
	lsconn.Write([]byte(cmd))
	//6.get resp from localserver
	var resp [1]byte
	lsconn.Read(resp[0:1])
	//7.go send data to localserver
	go transdata(conn, lsconn)
	//8.go get data from localserver
	go transdata(lsconn, conn)
}

func transdata(r, w net.Conn) {
	const bufmax uint = 1 << 20
	var buf [bufmax]byte
	for {
		n, err := r.Read(buf[0:bufmax])
		if err != nil {
			fmt.Println(err)
			break
		}
		n, err = w.Write(buf[0:n])
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
