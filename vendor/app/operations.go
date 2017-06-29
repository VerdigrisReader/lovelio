package app

import (
	"bytes"
	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
	"strconv"
)

// Types
type BoardItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type Board struct {
	BoardId string      `json:"board_id"`
	Name    string      `json:"name"`
	Items   []BoardItem `json:"items"`
}

type BoardName struct {
	BoardId string `json:"board_id"`
	Name    string `json:"name"`
}

type BoardList struct {
	Boards []Board `json:"boards"`
}

// NewUser creates a new user with a randomly generated uuid
// The user will eventualy be represented by a hash containing name and ID for each board
// It returns the uuid for the created user
func NewUser(conn redis.Conn) string {
	user := uuid.New()
	useruuid := user.String()
	return useruuid
}

// NewBoard creates a new board for the user
// A board is referenced in the user's Hash by (board_key str => name str)
// A board is a hash containing (name str => count int)
func NewBoard(conn redis.Conn, useruuid string) BoardName {
	var buffer bytes.Buffer

	count, _ := redis.Int64(conn.Do("HLEN", useruuid))
	stringKey := strconv.Itoa(int(count))

	buffer.WriteString(useruuid)
	buffer.WriteString(":")
	buffer.WriteString(stringKey)
	BoardId := buffer.String()

	_, err := conn.Do("HSET", useruuid, BoardId, "new")
	if err != nil {
		panic(err)
	}

	conn.Do("HSET", BoardId, "new", 0)
	return BoardName{BoardId, "new"}
}

// GetUserBoards returns a map of all user boards
// a board is defined in the user map containing (board_key => name)
// board_key must be unique
func GetUserBoards(conn redis.Conn, useruuid string) []BoardName {
	values, err := redis.StringMap(conn.Do("HGETALL", useruuid))
	var boards []BoardName

	if err != nil {
		return boards
	} else {
		for key, name := range values {
			boards = append(boards, BoardName{key, name})
		}
		return boards
	}
}

// GetBoardItems returns a map of board items
// a board is a map containing items and incremental counts
// board_key must be unique
func GetBoardItems(conn redis.Conn, BoardId string) []BoardItem {
	values, err := redis.Int64Map(conn.Do("HGETALL", BoardId))
	var items []BoardItem

	if err != nil {
		return items
	} else {
		for name, value := range values {
			items = append(items, BoardItem{name, value})
		}
		return items
	}
}

// RenameBoard, Given a BoardId and new name, renames the board to the passed value
// key must already exist
// Names can collide or be set multiple times
func RenameBoard(conn redis.Conn, useruuid, BoardId, newName string) {
	conn.Do("HSET", useruuid, BoardId, newName)
}

// IncrementBoardItem increases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns incremented key
func IncrementBoardItem(conn redis.Conn, BoardId, itemKey string) int64 {
	newValue, _ := redis.Int64(conn.Do("HINCRBY", BoardId, itemKey, 1))
	return newValue
}

// DecrementBoardItem decreases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns decremented key (must be >= 0)
func DecrementBoardItem(conn redis.Conn, BoardId, itemKey string) int64 {
	newValue, _ := redis.Int64(conn.Do("HINCRBY", BoardId, itemKey, -1))
	if newValue < 0 {
		conn.Do("HSET", BoardId, itemKey, 0)
		return 0
	} else {
		return newValue
	}
}

// RenameBoardItem moves the value of an existing board item to a new name within the hash
// no return value
func RenameBoardItem(conn redis.Conn, BoardId, itemKey, newName string) {
	currentValue, err := redis.Int64(conn.Do("HGET", BoardId, itemKey))
	if err == nil {
		conn.Do("HSET", BoardId, newName, currentValue)
		conn.Do("HDEL", BoardId, itemKey)
	}
}
