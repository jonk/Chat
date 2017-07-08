package main

import (
        "log"
        "net/http"
        "github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) //connected Clients map
var broadcast = make(chan Message) //broadcast channel

//this is used to take a normal http connection and upgrade it to a websocket
var upgrader = websocket.Upgrader{}

type Message struct {
    Email    string `json:"email"`
    Username string `json:"username"`
    Message  string `json:"message"`
}

func main() {
    //simple file server
    fs := http.FileServer(http.Dir("dist"))
    http.Handle("/", fs)
    http.HandleFunc("/ws", handleConnections)

    // fork a fancy go thread to handle incoming messages
    go handleMessages()

    // Start the server on localhost port 8000 and log any errors
    log.Println("http server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    //Upgrade GET request to a websocket
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }

    // this means closing will always happen
    defer ws.Close()

    // add it to our good ol' client map
    clients[ws] = true

    for {
        var msg Message
        // Read in a new message as JSON and map it to a Message object

        err := ws.ReadJSON(&msg)
        if err != nil {
            log.Printf("error: %v", err)
            delete(clients, ws)
            break
        }
        // send the new message to the broadcast channel
        broadcast <- msg
    }
}

func handleMessages() {
    for { 
        msg := <-broadcast

        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
                log.Printf("error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}
