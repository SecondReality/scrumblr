package main

import (
	"github.com/alicebob/miniredis"
	"fmt"
	"log"
	"encoding/json"
)

const redisPrefix = "GOSCRUMBLR"

type DB struct {
	redisClient *miniredis.Miniredis
}

func NewDB() *DB {
	tempdb := new(DB)
	if redisClient, err := miniredis.Run(); err != nil {
		fmt.Println("Error Initializing Database")
		fmt.Println("CATASTROPHIC FAILURE!")
	} else {
		fmt.Println("Database initialized")
		tempdb.redisClient = redisClient
	}

	return tempdb
}

// TODO: Make these functions have callbacks.
// Make an interface so there can be different implementations of the DB
// Make generic JSON conversion functions that are easier to use.
func (mydb *DB) SetTheme(room string, theme string) {
	mydb.redisClient.Set(redisPrefix+"-room:"+room+"-theme", theme)
}

func (mydb *DB) GetTheme(room string) string {
	theme, _ := mydb.redisClient.Get(redisPrefix + "-room:" + room + "-theme")
	return theme
}

func (mydb *DB) CreateColumn(room string, name string) {
	mydb.redisClient.Push(redisPrefix+"-room:"+room+"-columns", name)
}

func (mydb *DB) GetAllColumns(room string) []string {
	columns, _ := mydb.redisClient.List(redisPrefix + "-room:" + room + "-columns")
	return columns
}

func (mydb *DB) GetAllCards(room string) []interface{} {
	cardKeys, _ := mydb.redisClient.HKeys(redisPrefix + "-room:" + room + "-cards")

	var cards = make([]interface{}, len(cardKeys))
	for i, cardKey := range cardKeys {
		cardJson := mydb.redisClient.HGet(redisPrefix+"-room:"+room+"-cards", cardKey)
		var card map[string]interface{}
		if err := json.Unmarshal([]byte(cardJson), &card); err != nil {
			panic(err)
		} else {
			cards[i] = card
		}
	}

	return cards
}

func (mydb *DB) CreateCard(room string, id string, card interface{}) {

	if jsonBytes, err := json.Marshal(card); err != nil {
		log.Println("Error converting JSON")
	} else {
		jsonString := string(jsonBytes[:])
		mydb.redisClient.HSet(redisPrefix+"-room:"+room+"-cards", id, jsonString)
	}
}

func (mydb *DB) CardEdit(room string, id string, text string) {
	cardJson := mydb.redisClient.HGet(redisPrefix+"-room:"+room+"-cards", id)
	var card map[string]interface{}
	if err := json.Unmarshal([]byte(cardJson), &card); err != nil {
		panic(err) // TODO: Don't do this
	} else {
		card["text"] = text

		if jsonBytes, err := json.Marshal(card); err != nil {
			log.Println("Error converting JSON")
		} else {
			jsonString := string(jsonBytes[:])
			mydb.redisClient.HSet(redisPrefix+"-room:"+room+"-cards", id, jsonString)
		}
	}
}

func (mydb *DB) CardSetXY(room string, id string, x string, y string) {
	cardJson := mydb.redisClient.HGet(redisPrefix+"-room:"+room+"-cards", id)
	var card map[string]interface{}
	if err := json.Unmarshal([]byte(cardJson), &card); err != nil {

		panic(err) // TODO: Don't do this
	} else {
		card["x"] = x
		card["y"] = y
	}

	if jsonBytes, err := json.Marshal(card); err != nil {
		log.Println("Error converting JSON")
	} else {
		jsonString := string(jsonBytes[:])
		mydb.redisClient.HSet(redisPrefix+"-room:"+room+"-cards", id, jsonString)
	}
}

func (mydb *DB) SetBoardSize(room string, size interface{}) {
	if jsonBytes, err := json.Marshal(size); err != nil {
		log.Println("Error converting JSON")
	} else {
		jsonString := string(jsonBytes[:])
		mydb.redisClient.Set(redisPrefix+"-room:"+room+"-size", jsonString)
	}
}

func (mydb *DB) GetBoardSize(room string) map[string]interface{} {
	sizeJson, err := mydb.redisClient.Get(redisPrefix + "-room:" + room + "-size")

	if err != nil {
		return nil
	}

	var size map[string]interface{}
	if err := json.Unmarshal([]byte(sizeJson), &size); err != nil {
		panic(err)
	}

	return size
}
