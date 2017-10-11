package main

import (
	"github.com/deckarep/golang-set"
	"github.com/graarh/golang-socketio"
)

// Note: I'm fairly certain there's a huge bug in this code. userName variable is never set to anything, it's always undefined in the original Javascript.

// Dict mapping room names with people to sets of client objects.
// TODO: Is there no way to make a strongly typed map?
var rooms = make(map[string]mapset.Set) // Maps of *gosocketio.Channel
// Dict mapping sids to sets of rooms.
var sidRooms = make(map[string]mapset.Set)
// Dict mapping room names with people to sets of usernames.
var roomUsers = make(map[string]mapset.Set)

// Add a client to a room and return the sid:client mapping.
func AddToRoom(channel *gosocketio.Channel, room string, _ string, fn func([]interface{})) {
	if _, ok := sidRooms[channel.Id()]; !ok {
		sidRooms[channel.Id()] = mapset.NewSet()
	}
	sidRooms[channel.Id()].Add(room)

	if _, ok := rooms[room]; !ok {
		rooms[room] = mapset.NewSet()
	}

	rooms[room].Add(channel)

	if _, ok := roomUsers[room]; !ok {
		roomUsers[room] = mapset.NewSet()
	}
	roomUsers[room].Add(nil) // <SMR> Replicating a bug that was in the original code?

	fn(rooms[room].ToSlice())
}

func RoomsGetRoom(channel *gosocketio.Channel) string {
	if _, ok := sidRooms[channel.Id()]; !ok {
		return ""
	}
	return sidRooms[channel.Id()].ToSlice()[0].(string)
}

func AddToRoomAndAnnounce(channel *gosocketio.Channel, room string, userName string, msg interface{}) {
	AddToRoom(channel, room, userName, func(userNames []interface{}) {
		for _, client := range userNames {
			otherChannel := client.(*gosocketio.Channel)
			if otherChannel.Id() != channel.Id() {
				otherChannel.Emit("message", msg)
			}
		}
	})
}

// Return list of clients in the given room.
func RoomClients(room string) []*gosocketio.Channel {
	if roomSet, ok := rooms[room]; ok {
		// TODO: Look for a better way to convert a []interface{} to a []string.
		// TODO: Why doesn't the mapset have a length??
		roomSetSlice := roomSet.ToSlice()
		roomList := make([]*gosocketio.Channel, 0, len(roomSetSlice))
		for _, i := range roomSetSlice {
			roomList = append(roomList, i.(*gosocketio.Channel))
		}
		return roomList
	} else {
		return make([]*gosocketio.Channel, 0)
	}
}

func BroadcastToRoomates(c *gosocketio.Channel, action string, data interface{}) {
	// Build the message to deliver:
	msg := map[string]interface{}{
		"action": action,
		"data":   data,
	}

	roommates := mapset.NewSet()
	if _, ok := sidRooms[c.Id()]; ok {
		clientRoomsList := sidRooms[c.Id()].ToSlice()
		for _, room := range clientRoomsList {
			if _, ok := rooms[room.(string)]; ok {
				thisRoom := rooms[room.(string)].ToSlice()
				for _, thisParticularRoom := range thisRoom { // This is nasty - cyclomatic complexity is too damn high
					roommates.Add(thisParticularRoom)
				}
			}
		}
	}
	roommates.Remove(c)

	roommatesArray := roommates.ToSlice()
	for _, roommate := range roommatesArray {
		roommate.(*gosocketio.Channel).Emit("message", msg)
	}
}