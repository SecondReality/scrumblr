package main

import (
	"html/template"
	"log"
	"fmt"
	"net/http"
	"strings"
	"github.com/Joker/jade"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

var chttp = http.NewServeMux()
var server *gosocketio.Server
var userNames = make(map[*gosocketio.Channel]string)
var db = NewDB()

type Message struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

func getRoom(c *gosocketio.Channel, callback func(string)) {
	room := RoomsGetRoom(c)
	callback(room)
}

// Finish implementation of this (could do bad things without sanitization)
// (it's a place holder right now)
func scrub(text string) string {
	return text
}

func joinRoom(c *gosocketio.Channel, room string) {

	_, ok := userNames[c]
	if !ok {
		//  Since users join rooms before they set their name, set their name here if it wasn't previously set (javascript relies on string conversion to 'undefined')
		userNames[c] = "undefined"
	}

	msg := map[string]interface{}{
		"action": "join-announce",
		"data": map[string]interface{}{
			"sid":       "1",
			"user_name": userNames[c],
		},
	}

	AddToRoomAndAnnounce(c, room, userNames[c], msg)
}

// Not in original code base.
func clientJsonSend(c *gosocketio.Channel, action string, data interface{}) {
	msg := map[string]interface{}{
		"action": action,
		"data":   data,
	}
	c.Emit("message", msg)
}

func BroadcastToRoom(c *gosocketio.Channel, action string, data interface{}) {
	BroadcastToRoomates(c, action, data)
}

func ToString(val float64) string {
	return fmt.Sprintf("%f", val)
}

func main() {
	cleanAndInitializeDemoRoom()

	server = gosocketio.NewServer(transport.GetDefaultWebsocketTransport())
	if server != nil {
		log.Println("Server initialized!")
	}

	// Handle messages from the client:
	server.On("message", func(c *gosocketio.Channel, msg Message) string {

		if msg.Action == "" {
			return "OK"
		}

		switch msg.Action {
		case "joinRoom":
			{
				joinRoom(c, msg.Data.(string))
				c.Emit("message", map[string]interface{}{"action": "roomAccept", "data": ""})
			}
		case "initializeMe":
			{
				InitClient(c)
			}
			// TODO: Any easier way to get members from JSON? Don't do this at home folks. Maybe make a custom type? or the pointers in the struct https://blog.golang.org/json-and-go
			// BAD:     "top": scrub(msg.Data.(map[string]interface{})["position"].(string)),
		case "moveCard":
			{
				positionData := msg.Data.(map[string]interface{})
				cardID := positionData["id"].(string)
				cardLeft := positionData["position"].(map[string]interface{})["left"].(float64)
				cardRight := positionData["position"].(map[string]interface{})["top"].(float64)

				data := map[string]interface{}{
					"id": scrub(cardID),
					"position": map[string]interface{}{
						"left": ToString(cardLeft),
						"top":  ToString(cardRight),
					},
				}

				BroadcastToRoom(c, "moveCard", data)

				getRoom(c, func(room string) {
					db.CardSetXY(room, cardID, ToString(cardLeft), ToString(cardRight))
				})
			}

		case "createCard":
			{
				// TODO scrub/sanitize the card before sending, I'm lazy.
				// TODO: the non db create card function should be used
				data := msg.Data.(map[string]interface{})
				cardID := data["id"].(string)
				getRoom(c, func(room string) {
					db.CreateCard(room, cardID, data)
				})
				BroadcastToRoom(c, "createCard", data)
			}
		case "editCard":
			{
				data := msg.Data.(map[string]interface{})
				cardID := data["id"].(string)
				cardValue := data["value"].(string)

				getRoom(c, func(room string) {
					db.CardEdit(room, cardID, cardValue)
				})

				cleanData := map[string]interface{}{
					"value": cardValue,
					"id":    cardID,
				}

				BroadcastToRoom(c, "editCard", cleanData)
			}

		}
		return "OK"
	})

	chttp.Handle("/", http.FileServer(http.Dir("./client")))

	http.HandleFunc("/", HomeHandler)
	http.ListenAndServe(":8080", nil)
}

func InitClient(c *gosocketio.Channel) {
	onGetRoom := func(room string) {

		clientJsonSend(c, "initCards", db.GetAllCards(room))
		clientJsonSend(c, "initColumns", db.GetAllColumns(room))
		theme := db.GetTheme(room)
		// TODO: Verify if theme can be empty
		if theme == "" {
			theme = "bigcards"
		}
		clientJsonSend(c, "changeTheme", theme)

		// TODO: Verify if size can actually be nil
		if size := db.GetBoardSize(room); size != nil {
			clientJsonSend(c, "setBoardSize", size)
		}

		roomMatesClients := RoomClients(room)
		var roomMates = make([]map[string]interface{}, 0, 0)
		for _, roomMateClient := range roomMatesClients {
			if roomMateClient.Id() != c.Id() {
				newRoomMate := map[string]interface{}{
					"sid": roomMateClient.Id(),
					// This line is sketchy
					"user_name": userNames[roomMateClient],
				}
				roomMates = append(roomMates, newRoomMate)
			}
		}

		clientJsonSend(c, "initialUsers", roomMates)

	}
	getRoom(c, onGetRoom)
}

func createCard(room string, id string, text string, x float64, y float64, rot float64, colour string) {
	card := map[string]interface{}{
		"id":      id,
		"colour":  colour,
		"rot":     rot,
		"x":       x,
		"y":       y,
		"text":    text,
		"sticker": nil,
	}
	db.CreateCard(room, id, card)
}

func cleanAndInitializeDemoRoom() {
	db.CreateColumn("/demo", "Started")
	db.CreateColumn("/demo", "In Progress")
	db.CreateColumn("/demo", "Finished")
	createCard("/demo", "card1", "Hello this is fun", 300, 150, 0.5*10-5, "yellow")
}

//-------------------------------------------------------------

type TemplateData struct {
	Connected string
	Url       string
}

func HomePage(w http.ResponseWriter, _ *http.Request) {

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

func ScrumblrPage(w http.ResponseWriter, _ *http.Request) {

	layout, err := jade.ParseFile("views/layout.jade")
	if err != nil {
		log.Printf("\nParseFile error: %v", err)
	}

	index, err := jade.ParseFile("views/index.jade")
	if err != nil {
		log.Printf("\nParseFile error: %v", err)
	}

	temple := template.New("layout")
	goTpl, err := temple.Parse(layout)

	if err != nil {

		log.Printf("\nTemplate parse error: %v", err)
	}
	goTpl.New("index").Parse(index)

	err = goTpl.Execute(w, "")
	if err != nil {
		log.Printf("\nExecute error: %v", err)
	}

}

func HomeHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.URL.Path, ".") {

		if r.URL.Path == "/socket.io/socket.io.js" {
			chttp.ServeHTTP(w, r)
		} else

		if strings.Contains(r.URL.Path, "/socket.io/") {
			server.ServeHTTP(w, r)
		} else {
			chttp.ServeHTTP(w, r)
		}
	} else {
		if r.URL.Path == "/" {
			HomePage(w, r)
		} else {
			ScrumblrPage(w, r)
		}
	}
}
