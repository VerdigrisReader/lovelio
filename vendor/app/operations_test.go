package app_test

import (
	. "app"
	"github.com/rafaeljusto/redigomock"
	"regexp"
	"testing"
)

const TEST_UUID string = "051f2304-cb1c-40a1-9dec-97c02b879f4d"

// Test that the new user function returns a valid 36 char v4 uuid
func TestNewUser(t *testing.T) {
	conn := redigomock.NewConn()
	userid := NewUser(conn)

	r := regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")
	if !(r.MatchString(userid)) {
		t.Errorf("Expected a uuid, got %s", userid)
	}
}

func TestNewBoard(t *testing.T) {
	var expected_key string = "051f2304-cb1c-40a1-9dec-97c02b879f4d:0"

	conn := redigomock.NewConn()
	cmd := conn.Command("HLEN", TEST_UUID).Expect("0")
	addBoard := conn.Command("HSET", TEST_UUID, expected_key, "new")
	addFirstKey := conn.Command("HSET", expected_key, "new", 0)

	newBoardKey := NewBoard(conn, TEST_UUID)

	// Assert that index is concatenated
	if newBoardKey != expected_key {
		t.Errorf("Board key is not valid: %s", newBoardKey)
	}
	// Assert len is called
	if conn.Stats(cmd) != 1 {
		t.Errorf("HLEN was not called")
	}
	if conn.Stats(addBoard) != 1 {
		t.Errorf("HSET was not called to add board")
	}
	if conn.Stats(addFirstKey) != 1 {
		t.Errorf("HSET was not called to add new key")
	}
}

func TestGetUserBoards(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("HGETALL", TEST_UUID).ExpectMap(map[string]string{
		"key": "value",
	})

	var all_boards map[string]string
	all_boards = GetUserBoards(conn, TEST_UUID)

	if len(all_boards) != 1 {
		t.Errorf("Expected map with 1 value, got %s", all_boards)
	}

	if conn.Stats(cmd) != 1 {
		t.Errorf("HGETALL was not called")
	}
}

func TestGetBoardItems(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("HGETALL", TEST_UUID).ExpectMap(map[string]string{
		"key":  "232",
		"brey": "263",
	})

	var all_items map[string]int64
	all_items = GetBoardItems(conn, TEST_UUID)

	if len(all_items) != 2 {
		t.Errorf("Expected map with 2 values, got %s", all_items)
	}

	if conn.Stats(cmd) != 1 {
		t.Errorf("HGETALL was not called")
	}
}

func TestRenameBoard(t *testing.T) {
	conn := redigomock.NewConn()
	cmd := conn.Command("HSET", TEST_UUID, "key", "newname")

	RenameBoard(conn, TEST_UUID, "key", "newname")

	if conn.Stats(cmd) != 1 {
		t.Errorf("HSET was not called")
	}
}

func TestRenameBoardItem(t *testing.T) {
	conn := redigomock.NewConn()
	var expected int64 = 23
	cmd := conn.Command("HGET", TEST_UUID, "oldkey").Expect(expected)
	setcmd := conn.Command("HSET", TEST_UUID, "newkey", expected)
	delcmd := conn.Command("HDEL", TEST_UUID, "oldkey")

	RenameBoardItem(conn, TEST_UUID, "oldkey", "newkey")

	if conn.Stats(cmd) != 1 {
		t.Errorf("HSET was not called")
	}
	if conn.Stats(setcmd) != 1 {
		t.Errorf("HSET was not called")
	}
	if conn.Stats(delcmd) != 1 {
		t.Errorf("HSET was not called")
	}
}

func TestIncrementBoardItem(t *testing.T) {
	conn := redigomock.NewConn()
	var expected int64 = 1
	cmd := conn.Command("HINCRBY", TEST_UUID, "item", 1).Expect(expected)

	newValue := IncrementBoardItem(conn, TEST_UUID, "item")

	if conn.Stats(cmd) != 1 {
		t.Errorf("HSET was not called")
	}
	if newValue != 1 {
		t.Errorf("Expected incremented value 1, got %d, %v", newValue)
	}
}

func TestDecrementBoardItem(t *testing.T) {
	testcases := map[int64]int64{
		5:  5,
		-1: 0,
		1:  1,
		0:  0,
	}
	for input := range testcases {
		conn := redigomock.NewConn()
		cmd := conn.Command("HINCRBY", TEST_UUID, "item", -1).Expect(input)

		newValue := DecrementBoardItem(conn, TEST_UUID, "item")

		if conn.Stats(cmd) != 1 {
			t.Errorf("HSET was not called")
		}
		if newValue != testcases[input] {
			t.Errorf("Expected incremented value 1, got %d, %v", newValue)
		}
	}
}
