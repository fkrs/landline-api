package models

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	"bytes"
)

func TestDeleteRoom(t *testing.T) {
	room := makeFakeRoom()
	room.Id = "TestDeleteRoom-1"
	room.Slug = "TestDeleteRoom-slug1"
	room.TeamId = "TestDeleteRoom-team1"
	_ = insertFakeRoom(room, t)

	err := DeleteRoom(room.Slug, room.TeamId)
	if err != nil {
		t.Error("TestDeleteRoom error:", err)
	}

	result := Room{}
	err = Db.SelectOne(&result, "select * from rooms where id=$1", room.Id)
	if err != nil {
		t.Error("TestDeleteRoom error:", err)
	}
	if result.DeletedAt == nil {
		t.Error("TestDeleteRoom: DeletedAt should be updated (%+v)", result)
	}
	result.setTime(room.CreatedAt)
	if *room != result {
		t.Errorf("TestDeleteRoom: got (%+v), wanted (%+v)", result, *room)
	}
}

func TestFindOrCreateRoom(t *testing.T) {
	room := makeFakeRoom()
	room.Id = "TestFindOrCreateRoom-1"
	room.Slug = "TestFindOrCreateRoom-slug1"
	room.TeamId = "TestFindOrCreateRoom-team1"

	_, err := FindOrCreateRoom(room)
	if err != nil {
		t.Error("TestFindOrCreateRoom error:", err)
	}

	result := Room{}
	err = Db.SelectOne(&result, "select * from rooms where id=$1", room.Id)
	if err != nil {
		t.Error("TestFindOrCreateRoom error:", err)
	}
	result.setTime(room.CreatedAt)
	if *room != result {
		t.Errorf("TestFindOrCreateRoom: got (%+v), wanted (%+v)", result, *room)
	}
}

func TestFindRoom(t *testing.T) {
	room := makeFakeRoom()
	room.Id = "TestFindRoom-1"
	room.Slug = "TestFindRoom-slug1"
	room.TeamId = "TestFindRoom-team1"
	_ = insertFakeRoom(room, t)

	result, err := FindRoom(room.Slug, room.TeamId)
	if err != nil {
		t.Error("TestFindRoom error:", err)
	}

	result.setTime(room.CreatedAt)
	if *room != *result {
		t.Errorf("TestFindRoom: got (%+v), wanted (%+v)", result, room)
	}
}

func TestFindRooms(t *testing.T) {
	teamId := "TestFindRooms-team"
	rooms := []*Room{makeFakeRoom(), makeFakeRoom(), makeFakeRoom()}
	for i, room := range rooms {
		room.Id = fmt.Sprintf("TestFindRooms-%d", i)
		room.Slug = fmt.Sprintf("TestFindRooms-slug%d", i)
		room.TeamId = teamId
		_ = insertFakeRoom(room, t)
	}

	result, err := FindRooms(teamId)
	if err != nil {
		t.Error("TestFindRooms error:", err)
	}
	if len(result) != len(rooms) {
		t.Fatalf("TestFindRooms result length: got %d, want %d", len(result), len(rooms))
	}
	for i, room := range result {
		room.setTime(rooms[i].CreatedAt)
		if *rooms[i] != room {
			t.Errorf("TestFindRooms: got (%+v), want (%+v)", room, *rooms[i])
		}
	}
}

func TestFindRoomById(t *testing.T) {
	room := makeFakeRoom()
	room.Id = "TestFindRoomById-1"
	_ = insertFakeRoom(room, t)

	result := FindRoomById(room.Id)

	result.setTime(room.CreatedAt)
	if *room != *result {
		t.Errorf("TestFindRoomById: got (%+v), wanted (%+v)", result, room)
	}
}

func TestUpdateRoom(t *testing.T) {
	room := makeFakeRoom()
	room.Id = "TestUpdateRoom-1"
	room.TeamId = "TestUpdateRoom-team1"
	room.Slug = "TestUpdateRoom-slug1"
	_ = insertFakeRoom(room, t)

	fields := makeFakeRoom()
	fields.Slug = "TestUpdateRoom-slug2"
	fields.Topic = "TestUpdateRoom-topic2"

	result, err := UpdateRoom(room.Slug, room.TeamId, fields)
	if err != nil {
		t.Fatal("TestUpdateRoom:", err)
	}
	room.Slug = fields.Slug
	room.Topic = fields.Topic
	result.setTime(room.CreatedAt)
	if *room != *result {
		t.Errorf("TestUpdateRoom: got (%v), want (%v)", result, room)
	}
}

func TestSubscribers(t *testing.T) {
	membership := makeFakeRoomMembership()
	membership.Id = "TestSubscribers-1"
	membership.RoomId = "TestSubscribers-room1"
	membership.UserId = "TestSubscribers-user1"
	_ = insertFakeRoomMembership(membership, t)

	result, err := Subscribers(membership.RoomId)
	if err != nil {
		t.Fatal("TestSubscribers error:", err)
	}
	if len(*result) != 1 {
		t.Fatalf("TestSubscribers got %d results, want %d results", len(*result), 1)
	}
	if (*result)[0] != membership.UserId {
		t.Errorf("TestSubscribers got %s, want %s", (*result)[0], membership.UserId)
	}
}

func TestUnreadRooms(t *testing.T) {
	userId := "TestUnreadRooms-user"
	expected := "TestUnreadRooms-response"
	response := fmt.Sprintf(`{"response":"%s"}`, expected)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != "GET" {
			t.Errorf("TestUnreadRooms: got (%s), want (%s)", r.Method, "GET")
		}
		if r.RequestURI != "/readers/" + userId {
			t.Errorf("TestUnreadRooms: got (%s), want (%s)", r.RequestURI, "/articles")
		}
		if r.Header.Get("Authorization") != "Basic eHl6Og==" {
			t.Errorf("TestUnreadRooms: got (%s), want (%s)", r.Header.Get("Authorization"), "Basic eHl6Og==")
		}
		w.Write(bytes.NewBufferString(response).Bytes())
	}))
	defer server.Close()
	err := os.Setenv("RR_URL", server.URL)
	if err != nil {
		t.Fatal("TestUnreadRooms error:", err)
	}
	err = os.Setenv("RR_PRIVATE_KEY", "xyz")
	if err != nil {
		t.Fatal("TestUnreadRooms error:", err)
	}

	resp, err := UnreadRooms(userId)
	if err != nil {
		t.Fatal("TestUnreadRooms error:", err)
	}
	result := resp.(map[string]interface{})
	if result["response"] != expected {
		t.Errorf("TestUnreadRooms got %s, want %s", result["response"], expected)
	}
}

func insertFakeRoom(room *Room, t *testing.T) *Room {
	err := Db.Insert(room)
	if err != nil {
		t.Fatal("Insert fake room error:", err)
	}
	return room
}

func (o *Room) setTime(t time.Time) {
	o.CreatedAt = t
	o.UpdatedAt = t
	o.DeletedAt = nil
}

func makeFakeRoom() *Room {
	t := time.Now()
	return &Room{
		Id:        "1",
		CreatedAt: t,
		UpdatedAt: t,
		DeletedAt: nil,
		TeamId:    "teamid1",
		Slug:      "slug1",
		Topic:     "topic1",
	}
}
