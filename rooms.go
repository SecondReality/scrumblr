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

func BroadcastToRoomates(c *gosocketio.Channel, action string, data interface{}) {
	// Build the message to deliver:
	msg := map[string]interface{}{
			"action": action,
			"data": data,
	}

	roommates := mapset.NewSet()
	if _, ok := sid_rooms[c.Id()]; ok {
		clientRoomsList := sid_rooms[c.Id()].ToSlice()
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

/*
// Broadcast message to all the other clients that are in rooms with this client
exports.broadcast_to_roommates = function (client, msg) {
	var roommates = new sets.Set();

   if (sid_rooms.hasOwnProperty(client.id))
	{
		var client_rooms = sid_rooms[client.id].array();
		for (var i = 0; i < client_rooms.length; i++)
		{
		   var room = client_rooms[i];
		   if (rooms.hasOwnProperty(room))
			{
				var this_room = rooms[room].array();
				for (var j = 0; j < this_room.length; j++)
					roommates.add(this_room[j]);
		   }
		}
	}

	//remove self from the set
	roommates.remove(client);
	roommates = roommates.array();

	//console.log('client: ' + client.id + " is broadcasting to: ");


   for (var k = 0; k < roommates.length; k++)
	{
		//console.log('  - ' + roommates[i].id);
		roommates[k].json.send(msg);
	}
};
*/
