package main

	import (
		"github.com/deckarep/golang-set"
		"github.com/graarh/golang-socketio"
		"log"
	)
// Note: I'm fairly certain there's a huge bug in this code. userName variable is never set to anything, it's always undefined in the original Javascript.

// Dict mapping room names with people to sets of client objects.
// TODO: Is there no way to make a strongly typed map?
var rooms = make(map[string]mapset.Set) // Maps of *gosocketio.Channel
//mapset.NewSet()
//var sid_rooms =
// mapset.
// Dict mapping sids to sets of rooms.
var sid_rooms = make(map[string]mapset.Set)
// Dict mapping room names with people to sets of usernames.
var room_users = make(map[string]mapset.Set)

// Add a client to a room and return the sid:client mapping.
func AddToRoom(channel *gosocketio.Channel, room string, userName string, fn func([]interface{})){
	if _, ok := sid_rooms[channel.Id()]; !ok {
		sid_rooms[channel.Id()] = mapset.NewSet()
	}
	sid_rooms[channel.Id()].Add(room)

	if _, ok := rooms[room]; !ok {
		rooms[room] = mapset.NewSet()
	}

	rooms[room].Add(channel)
	log.Println("added username " + userName)

	if _, ok := room_users[room]; !ok {
		room_users[room] = mapset.NewSet()
	}
	room_users[room].Add(nil); // <SMR> Replicating a bug that was in the original code?

	fn(rooms[room].ToSlice())
}

func RoomsGetRoom(channel *gosocketio.Channel) string {
	if _, ok := sid_rooms[channel.Id()]; !ok {
		return ""
	}
	return sid_rooms[channel.Id()].ToSlice()[0].(string)
}

func AddToRoomAndAnnounce(channel *gosocketio.Channel, room string, userName string, msg interface{}){
	AddToRoom(channel, room, userName, func(userNames []interface{}){
		for _, client := range userNames {
			otherChannel := client.(*gosocketio.Channel)
			log.Println("adding to room " + otherChannel.Id())
			if(otherChannel==nil){
				log.Println("yeah null")
			}
			if(otherChannel.Id() != channel.Id()){
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
				    roomList = append(roomList, i.( *gosocketio.Channel))
				}
			return roomList
		} else {
			return make([]*gosocketio.Channel, 0)
		}
};

/*
exports.add_to_room_and_announce = function (client, room, msg) {

		// Add user info to the current dramatis personae
		exports.add_to_room(client, room, function(clients) {
		    // Broadcast new-user notification
		    for (var i = 0; i < clients.length; i++)
			{
				if (clients[i].id != client.id)
					clients[i].json.send(msg);
			}
		});
};
*/
