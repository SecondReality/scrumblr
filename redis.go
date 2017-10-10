package main

	import (
"github.com/alicebob/miniredis"
"fmt"
    "log"
    "encoding/json"
)

const redis_prefix = "GOSCRUMBLR"

type DB struct {
    redisClient * miniredis.Miniredis
}

func NewDB() *DB {
    tempdb := new(DB)
    if redisClient, err := miniredis.Run(); err!=nil {
            fmt.Println("Error Initializing Database")
            fmt.Println("CATASTROPHIC FAILURE!")
    } else {
    fmt.Println("Database initialized")
    tempdb.redisClient = redisClient
}

    return tempdb
}

/*
getAllCards: function(room, callback) {
    redisClient.hgetall(REDIS_PREFIX + '-room:' + room + '-cards', function (err, res) {

        var cards = [];

        for (var i in res) {
            cards.push( JSON.parse(res[i]) );
        }
        //console.dir(cards);

        callback(cards);
    });
},
*/

// HKEYS and HGET?

func (mydb * DB)CreateColumn(room string, name string){
    mydb.redisClient.Push(redis_prefix+ "-room:" + room + "-columns", name)
}

func (mydb * DB)GetAllColumns(room string) []string {
    columns, _ := mydb.redisClient.List(redis_prefix+ "-room:" + room + "-columns")
    return columns
}

func (mydb * DB)GetAllCards(room string) []interface{} {
    cardKeys, _ := mydb.redisClient.HKeys(redis_prefix+ "-room:" + room + "-cards")

    var cards = make([]interface{}, len(cardKeys))
    for i, cardKey := range cardKeys {
        cardJson := mydb.redisClient.HGet(redis_prefix+ "-room:" + room + "-cards", cardKey)
        var card map[string]interface{}
        if err := json.Unmarshal([]byte(cardJson), &card); err != nil {
        panic(err)
        } else {
            cards[i] = card
        }
    }

    return cards
}

func (mydb * DB)CreateCard(room string, id string, card interface{}){

    if jsonBytes, err := json.Marshal(card); err!=nil {
        log.Println("Error converting JSON")
    } else {
        jsonString := string(jsonBytes[:])
        mydb.redisClient.HSet(redis_prefix+ "-room:" + room + "-cards", id, jsonString)
    }
}

/*
createCard: function(room, id, card) {
    var cardString = JSON.stringify(card);
    redisClient.hset(
        REDIS_PREFIX + '-room:' + room + '-cards',
        id,
        cardString
    );
},
*/

