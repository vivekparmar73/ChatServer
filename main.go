package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/websocket"
)

var clients = make(map[string]*websocket.Conn)
var mutex = &sync.Mutex{}

func wsHandler(ws *websocket.Conn) {
	defer ws.Close()

	var username string
	err := websocket.Message.Receive(ws, &username)
	if err != nil {
		log.Println("Failed to read username:", err)
		return
	}

	mutex.Lock()
	clients[username] = ws
	mutex.Unlock()

	fmt.Printf("%s connected\n", username)

	for {
		var msg string
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Printf("%s disconnected\n", username)
			break
		}

		// Expected format: recipient|message
		parts := strings.SplitN(msg, "|", 2)
		if len(parts) != 2 {
			websocket.Message.Send(ws, "Invalid format. Use recipient|message")
			continue
		}

		toUser := parts[0]
		message := parts[1]

		mutex.Lock()
		receiverConn, ok := clients[toUser]
		mutex.Unlock()

		if ok {
			websocket.Message.Send(receiverConn, fmt.Sprintf("[%s]: %s", username, message))
		} else {
			websocket.Message.Send(ws, fmt.Sprintf("%s is not online", toUser))
		}
	}

	mutex.Lock()
	delete(clients, username)
	mutex.Unlock()
}

func main() {
	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		handler := websocket.Handler(wsHandler)
		handler.ServeHTTP(c.Writer, c.Request)
	})

	// Serve the HTML file at http://localhost:8080/
	r.StaticFile("/", "./index.html")

	fmt.Println("Server started at ws://localhost:8080/ws")
	r.Run(":8080")
}
