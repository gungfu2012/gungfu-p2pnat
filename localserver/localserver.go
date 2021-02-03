package main

import (
	//"bytes"
	"fmt"
	"flag"
	//"io/ioutil"
	//"log"
	"net"
	//"time"
	//"math/rand"
	"github.com/gorilla/websocket"
	//"net/http"
	//"net/http/httptrace"
	//"strconv"
)

var laddr string
var remoteserverpath string = "localserver"
var remoteserver string

func main() {
	//get localserver ipv6 addr
	Interface, _ := net.InterfaceByName("eth0")
	addrs, _ := Interface.Addrs()
	for _, addr := range addrs {
		ip, _, _ := net.ParseCIDR(addr.String())
		if ip.To4() == nil && ip.IsGlobalUnicast() {
			laddr = ip.String()
			fmt.Println(laddr)
			break
		}
	}
	//1.wss to remoteserver
	flag.StringVar(&remoteserver, "remoteserver", "ws://127.0.0.1:8080/", "default remote server")
	flag.Parse()
	wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+remoteserverpath, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.send localserver ipv6 to remortserver
	err = wsconn.WriteMessage(websocket.BinaryMessage, []byte(laddr))
	if err != nil {
		fmt.Println("send localserver ipv6 to remotrserver err : ", err)
		return
	}
	//3.loop for read client ipv6:port
	for {
		_, recvbuf, err := wsconn.ReadMessage()
		if err != nil {
			fmt.Println("read client ipv6:port err : ", err)
			continue
		}
		raddr := string(recvbuf)
		go handleconn(raddr)
	}
}

func handleconn(raddr string) {
	//4.dial to client
	_, port, _ := net.SplitHostPort(raddr)
	laddrp := "[" + laddr + "]:" + port
	laddrtcp, _ := net.ResolveTCPAddr("tcp6", laddrp)
	raddrtcp, _ := net.ResolveTCPAddr("tcp6", raddr)
	clientconn, err := net.DialTCP("tcp6", laddrtcp, raddrtcp)
	if err != nil {
		fmt.Println("dial to client err : ", err)
		return
	}
	//5.get cmd from client
	var buf [1024]byte
	n, err := clientconn.Read(buf[0:1024])
	if err != nil {
		fmt.Println("get cmd from client : ", err)
		return
	}
	appport := string(buf[0:n])
	//6.dial to local app
	laconn, err := net.Dial("tcp", "127.0.0.1:" + appport)
	if err != nil {
		fmt.Println("dial to local app err : ", err)
		return
	}
	//7.send resp to client
	var resp = [1]byte{0x01}
	n, err = clientconn.Write(resp[0:1])
	if err != nil {
		fmt.Println("send resp to client err : ", err)
		return
	}
	//8.go send data to localserver
	go transdata(laconn, clientconn)
	//9.go get data from localserver
	go transdata(clientconn, laconn)
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
