package main

import (
    "html/template"
    "log"
    "fmt"
    "net/http"
    "strings"
    "github.com/Joker/jade"
    "golang.org/x/net/websocket"
    "io"
    "github.com/graarh/golang-socketio"
    "github.com/graarh/golang-socketio/transport"
    "lib/rooms"
)

var chttp = http.NewServeMux()
var server * gosocketio.Server

func echoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

type Message struct {
  Action string `json:"action"`
  Data string `json:"data"`
}

type Channel struct {
	Channel string `json:"channel"`
}

type successFunction func()

func joinRoom(/*client,*/ room string, fn successFunction) {
  msg := map[string]interface{}{
      "action": "join-announce",
      "data": map[string]interface{}{
          "sid": "1",
          "user_name": "little bobby tables",
      },
  }

/*
  msg := &Message{
          Action:   "join-announce",
          Data: []string{"apple", "peach", "pear"}}
	msg.action = '';
	msg.data		= { sid: client.id, user_name: client.user_name };
*/
	//rooms.add_to_room_and_announce(client, room, msg);
	fn();
}

func main() {
  server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
  if(server!=nil){
    log.Println("ALL OIK!")
  }

  server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
  		log.Println("New client connected")
  	})

    server.On("joinRoom", func(c *gosocketio.Channel, channel string) string {

  		log.Println("Client joined to ", channel)
  		return "joined to " + channel
  	})

    // Handle messages from the client:
  	server.On("message", func(c *gosocketio.Channel, msg Message) string {

      log.Println(msg.Action)

      switch(msg.Action) {
        case "joinRoom":
          log.Println("Client Joined Room")
      }

      /*
      server.BroadcastToAll("my event", "What up")
      log.Println("CHANNEL IS " + c.Id())
      c.Emit("will this work?", "blah")
  		c.BroadcastTo(c.Id(), "message", "just testing")
      */

  		return "OK"
  	})

    server.On(gosocketio.OnDisconnection, func(c *gosocketio.Channel) {
    		log.Println("Disconnected")
    	})

    chttp.Handle("/", http.FileServer(http.Dir("./client")))

    http.HandleFunc("/", HomeHandler)
    http.ListenAndServe(":8080", nil)
}

type TemplateData struct {
    Connected string
    Url string
}

func HomePage(w http.ResponseWriter, r *http.Request){

	layout, err := jade.ParseFile("views/home.jade")
	if err != nil {
		log.Printf("\nParseFile error: %v", err)
	}

  // Insert template variables:
  goTpl, err := template.New("html").Parse(layout)
  if err != nil {
    fmt.Printf("\nTemplate parse error: %v", err)
  }

  p := TemplateData{Connected: "123", Url: "whatup"}
  err = goTpl.Execute(w, p)
	if err != nil {
		fmt.Printf("\nExecute error: %v", err)
		return
	}

}

func ScrumblrPage(w http.ResponseWriter, r *http.Request){

  layout, err := jade.ParseFile("views/layout.jade")
	if err != nil {
		log.Printf("\nParseFile error: %v", err)
	}

	index, err := jade.ParseFile("views/index.jade")
	if err != nil {
		log.Printf("\nParseFile error: %v", err)
	}

  temple := template.New("layout")
  go_tpl, err := temple.Parse(layout)

  	if err != nil {

  		log.Printf("\nTemplate parse error: %v", err)
  	}
    go_tpl.New("index").Parse(index)

  	err = go_tpl.Execute(w, "")
  	if err != nil {
  		log.Printf("\nExecute error: %v", err)
  	}

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

    if (strings.Contains(r.URL.Path, ".")) {

      if(r.URL.Path=="/socket.io/socket.io.js") {
        chttp.ServeHTTP(w, r)
      } else

      if (strings.Contains(r.URL.Path, "/socket.io/")) {
        server.ServeHTTP(w, r)
      } else {
        log.Printf("Serving a file")
        chttp.ServeHTTP(w, r)
      }
    } else {
        if r.URL.Path=="/" {
          HomePage(w, r)
        } else {
          ScrumblrPage(w, r)
        }
    }
}
