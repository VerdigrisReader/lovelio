package lovelio

import (
	"bytes"
	"github.com/garyburd/redigo/redis"
	"github.com/google/uuid"
)

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
func NewBoard(conn redis.Conn, useruuid string) string {
	var buffer bytes.Buffer

	stringKey, _ := redis.String(conn.Do("HLEN", useruuid))

	buffer.WriteString(useruuid)
	buffer.WriteString(":")
	buffer.WriteString(stringKey)
	boardKey := buffer.String()

	conn.Do("HSET", useruuid, boardKey, "new")
	conn.Do("HSET", boardKey, "new", 0)
	return boardKey
}

// GetUserBoards returns a map of all user boards
// a board is defined by a map containing (board_key => name)
// board_key must be unique
func GetUserBoards(conn redis.Conn, useruuid string) map[string]string {
	values, err := redis.StringMap(conn.Do("HGETALL", useruuid))

	if err != nil {
		var blank map[string]string
		return blank
	} else {
		return values
	}
}

// RenameBoard, Given a boardKey and new name, renames the board to the passed value
// key must already exist
// Names can collide or be set multiple times
func RenameBoard(conn redis.Conn, useruuid, boardKey, newName string) {
	conn.Do("HSET", useruuid, boardKey, newName)
}

// IncrementBoardItem increases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns incremented key
func IncrementBoardItem(conn redis.Conn, boardKey, itemKey string) int64 {
	newValue, _ := redis.Int64(conn.Do("HINCRBY", boardKey, itemKey, 1))
	return newValue
}

// DecrementBoardItem decreases the key by 1
// If the item doesn't exist the value is set to 1, so this can be used to create a new item
// Returns decremented key (must be >= 0)
func DecrementBoardItem(conn redis.Conn, boardKey, itemKey string) int64 {
	newValue, _ := redis.Int64(conn.Do("HINCRBY", boardKey, itemKey, -1))
	if newValue < 0 {
		conn.Do("HSET", boardKey, itemKey, 0)
		return 0
	} else {
		return newValue
	}
}

// RenameBoardItem moves the value of an existing board item to a new name within the hash
// no return value
func RenameBoardItem(conn redis.Conn, boardKey, itemKey, newName string) {
	currentValue, err := redis.Int64(conn.Do("HGET", boardKey, itemKey))
	if err != nil {
		conn.Do("HSET", boardKey, newName, currentValue)
		conn.Do("HDEL", boardKey, itemKey)
	}
}
