package main

import (
	//"bytes"
	"flag"
	"fmt"
	//"io/ioutil"
	//"log"
	"net"
	"time"
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
	time.Sleep(30 * time.Second)
	//get localserver ipv6 addr
	Interface, err := net.InterfaceByName("eth0")
	if err != nil {
		fmt.Println("get localserver eth0 err : ", err)
		return
	}
	addrs, err := Interface.Addrs()
	if err != nil {
		fmt.Println("get localserver addrs err : ", err)
		return
	}
	for _, addr := range addrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			fmt.Println("get localserver ip err : ", err)
			return
		}
		if ip.To4() == nil && ip.IsGlobalUnicast() {
			laddr = ip.String()
			fmt.Println(laddr)
			break
		}
	}
	//1.wss to remoteserver
	flag.StringVar(&remoteserver, "remoteserver", "ws://127.0.0.1:8080/", "default remote server")
	flag.Parse()
	for {
		time.Sleep(10 * time.Second)
		wsconn, _, err := websocket.DefaultDialer.Dial(remoteserver+remoteserverpath, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		//2.send localserver ipv6 to remortserver
		if wsconn == nil {
			continue
		}
		err = wsconn.WriteMessage(websocket.BinaryMessage, []byte(laddr))
		if err != nil {
			fmt.Println("send localserver ipv6 to remotrserver err : ", err)
			continue
		}
		//3.loop for read client ipv6:port
		readerrcount := 0
		for {
			_, recvbuf, err := wsconn.ReadMessage()
			if err != nil {
				fmt.Println("read client ipv6:port err : ", err)
				readerrcount ++
				if readerrcount > 5 {
					break
				}
				continue
			}
			readerrcount = 0
			raddr := string(recvbuf)
			go handleconn(raddr)
		}
	}
}

func handleconn(raddr string) {
	//4.dial to client
	_, port, err := net.SplitHostPort(raddr)
	if err != nil {
		fmt.Println("splite host port err : ", err)
		return
	}
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
	if clientconn == nil {
		return
	}
	n, err := clientconn.Read(buf[0:1024])
	if err != nil {
		fmt.Println("get cmd from client : ", err)
		return
	}
	appport := string(buf[0:n])
	//6.dial to local app
	laconn, err := net.Dial("tcp", "127.0.0.1:"+appport)
	if err != nil {
		fmt.Println("dial to local app err : ", err)
		return
	}
	//7.send resp to client
	var resp = [1]byte{0x01}
	if clientconn == nil {
		return
	}
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
		if r == nil {
			return
		}
		n, err := r.Read(buf[0:bufmax])
		if err != nil {
			fmt.Println(err)
			break
		}
		if w == nil {
			return
		}
		n, err = w.Write(buf[0:n])
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}
