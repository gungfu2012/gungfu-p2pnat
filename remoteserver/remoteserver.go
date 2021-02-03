package main

import (
	//"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	//"strconv"
)

const bufmax uint = 1 << 20

var localserveraddr string
var localserverconn *websocket.Conn

var upgrader = websocket.Upgrader{}

func localserver(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	//1.save the localserver ws conn
	if err != nil {
		fmt.Println("localserver ws upgrade err : ", err)
		return
	}
	localserverconn = c
	//2.read and save localserver addr
	_, buf, err := localserverconn.ReadMessage()
	if err != nil {
		fmt.Println("ws read localserver ip err : ", err)
		return
	}
	localserveraddr = string(buf[0:len(buf)])
	fmt.Println("the local server ip is : ",localserveraddr)
}

func client(w http.ResponseWriter, r *http.Request) {
	if r == nil {
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("client ws upgrade err : ", err)
		return
	}
	//1.read client ipv6:port
	_, buf, err := c.ReadMessage()
	if err != nil {
		fmt.Println("read client ip err : ", err)
		return
	}
	//2.send client ipv6:port to localserver
	err = localserverconn.WriteMessage(websocket.BinaryMessage, buf)
	if err != nil {
		fmt.Println("send client ip to localserver err : ", err)
		return
	}
	//3.send localserver ip v6 to client
	err = c.WriteMessage(websocket.BinaryMessage, []byte(localserveraddr))
	if err != nil {
		fmt.Println("send localserver ip to client err : ", err)
		return
	}
}

func main() {
	http.HandleFunc("/localserver", localserver)
	http.HandleFunc("/client", client)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	fmt.Println("this server addr is : ", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
