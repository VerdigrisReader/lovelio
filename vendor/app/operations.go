package app

import (
	"bytes"
	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
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
	var buffer bytes.Buffer

	user := uuid.New()
	useruuid := user.String()

	buffer.WriteString("user:")
	buffer.WriteString(useruuid)
	return buffer.String()
}

func stringAppend(key, suffix string) string {
	var buffer bytes.Buffer
	buffer.WriteString(key)
	buffer.WriteString(suffix)
	return buffer.String()
}

// NewBoard creates a new board for the user
// A board is referenced in the user's Hash by (board_key str => name str)
// A board is a hash containing (name str => count int)
func NewBoard(conn redis.Conn, useruuid, boardName string) BoardName {
	newUUID := uuid.New().String()
	boardId := stringAppend("board:", newUUID)
	boardItemsId := stringAppend(boardId, ":items")

	conn.Send("RPUSH", useruuid, boardId)

	// Assign name to board struct
	conn.Send("HSET", boardId, "name", boardName)
	conn.Send("HSET", boardId, "itemsId", boardItemsId)
	// Add first member to boardItems sortedset
	conn.Send("ZADD", boardItemsId, 0, "new")
	err := conn.Flush()
	if err != nil {
		panic(err)
	}
	return BoardName{boardId, boardName}
}

func getManyBoards(conn redis.Conn, keys ...string) []BoardName {
	var boards []BoardName

	for _, key := range keys {
		conn.Send("HGET", key, "name")
	}
	conn.Flush()
	for _, key := range keys {
		name, _ := redis.String(conn.Receive())
		boards = append(boards, BoardName{key, name})
	}
	return boards
}

// GetUserBoards returns a map of all user boards
// a board is defined in the user map containing (board_key => name)
// board_key must be unique
func GetUserBoards(conn redis.Conn, useruuid string) []BoardName {
	values, err := redis.Strings(conn.Do("LRANGE", useruuid, 0, -1))
	var boards []BoardName

	if err != nil {
		return boards
	} else {
		return getManyBoards(conn, values...)
	}
}

// GetBoardItems returns a map of board items
// a board is a map containing items and incremental counts
// board_key must be unique
func GetBoardItems(conn redis.Conn, boardId string) []BoardItem {
	itemsId := stringAppend(boardId, ":items")
	values, err := redis.Int64Map(conn.Do("ZRANGE", itemsId, 0, -1, "WITHSCORES"))
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
func RenameBoard(conn redis.Conn, boardId, newName string) {
	conn.Do("HSET", boardId, "name", newName)
}

// IncrementBoardItem increases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns incremented key
func IncrementBoardItem(conn redis.Conn, boardId, itemKey string) int64 {
	itemsId := stringAppend(boardId, ":items")
	newValue, err := redis.Int64(conn.Do("ZINCRBY", itemsId, 1, itemKey))
	if err != nil {
		panic(err)
	}
	return newValue
}

// DecrementBoardItem decreases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns decremented key (must be >= 0)
func DecrementBoardItem(conn redis.Conn, boardId, itemKey string) int64 {
	itemsId := stringAppend(boardId, ":items")
	newValue, err := redis.Int64(conn.Do("ZINCRBY", itemsId, -1, itemKey))
	if err != nil {
		panic(err)
	}
	if newValue < 0 {
		conn.Do("ZADD", itemsId, 0, itemKey)
		return 0
	} else {
		return newValue
	}
}

// RenameBoardItem moves the value of an existing board item to a new name within the hash
// no return value
func RenameBoardItem(conn redis.Conn, boardId, itemKey, newName string) {
	itemsId := stringAppend(boardId, ":items")
	currentValue, err := redis.Int64(conn.Do("ZSCORE", itemsId, itemKey))
	if err == nil {
		conn.Do("ZADD", itemsId, newName, currentValue)
		conn.Do("ZREM", itemsId, itemKey)
	}
}
