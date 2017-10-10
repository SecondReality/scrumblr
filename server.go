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
)

var chttp = http.NewServeMux()
var server * gosocketio.Server
var userNames = make(map[*gosocketio.Channel]string)
var db = NewDB()

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

// This doesn"t exist in the original source - had to make this type in order to extend with username
type Client struct {

}

type successFunction func()

func getRoom(c *gosocketio.Channel, callback func(string)){
	room := RoomsGetRoom(c);
	callback(room);
}

func setUserName(c *gosocketio.Channel, userName string){
    userNames[c]=userName
}

func joinRoom(c *gosocketio.Channel, room string, fn successFunction) {

    _, ok := userNames[c]
    if(!ok){
    // <SMR> Since users join rooms before they set their name, set their name here if it wasn't previously set (javascript relies on string conversion to 'undefined')
    userNames[c]="undefined"
     }

  msg := map[string]interface{}{
      "action": "join-announce",
      "data": map[string]interface{}{
          "sid": "1",
          "user_name": userNames[c],
      },
  }

/*
  jsonBytes, err := json.Marshal(msg)
  if(err !=nil){
  fmt.Println(err)
  }
  jsonString := string(jsonBytes[:])
  */
/*
  msg := &Message{
          Action:   "join-announce",
          Data: []string{"apple", "peach", "pear"}}
	msg.action = "";
	msg.data		= { sid: client.id, user_name: client.user_name }
*/
	AddToRoomAndAnnounce(c, room, userNames[c], msg);
}

// Not in original code base.
func clientJsonSend(c *gosocketio.Channel, action string, data interface{}) {
    msg := map[string]interface{}{
        "action": action,
        "data": data,
    }
    c.Emit("message", msg)
}

func main() {
  cleanAndInitializeDemoRoom()
  //db.CreateCard("/demo", "potato", "cardcontent")
  //db.GetAllCards("/demo")
  server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
  if(server!=nil){
    log.Println("Server initialized!")
  }

  server.On(gosocketio.OnConnection, func(c *gosocketio.Channel) {
  		log.Println("New client connected")
  	})

    // Handle messages from the client:
  	server.On("message", func(c *gosocketio.Channel, msg Message) string {

      log.Println(msg.Action)

      switch(msg.Action) {
        case "joinRoom":
          log.Println("Client Joined Room")
          joinRoom(c, msg.Data, func(){})
          c.Emit("message", map[string]interface{}{ "action": "roomAccept", "data": "" } );
      case "initializeMe":
          InitClient(c)
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

func InitClient(c *gosocketio.Channel){
    onGetRoom := func(room string) {

        clientJsonSend(c, "initCards", db.GetAllCards(room))
        clientJsonSend(c, "initColumns", db.GetAllColumns(room))

        log.Println("got room")
        }
    getRoom(c, onGetRoom)
}


func createCard(room string, id string, text string, x float64, y float64, rot float64, colour string) {
    card := map[string]interface{}{
        "id": id,
        "colour": colour,
        "rot": rot,
        "x": x,
        "y": y,
        "text": text,
        "sticker": nil,
	};

	db.CreateCard(room, id, card);
}

func cleanAndInitializeDemoRoom(){
    db.CreateColumn("/demo", "Started" )
    createCard("/demo", "card1", "Hello this is fun", 300, 150, 0.5 * 10 - 5, "yellow");
}
//-------------------------------------------------------------
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
