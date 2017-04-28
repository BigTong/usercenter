package user

import (
	"encoding/json"
)

const (
	RELATION_STATE_LIKED    = "liked"
	RELATION_STATE_DISLIKED = "disliked"
	RELATION_STATE_MATCHED  = "matched"
)

type User struct {
	_id         int64
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	Createdtime int64  `json:"createdTime,omitempty"`
	Address     string `json:"address,omitempty"`
	Type        string `json:"type,omitempty"`
}

type UserRelationShip struct {
	Id        int64  `json:"id"`
	State     string `json:"state"`
	Otherside int64  `json:"other_side"`
}

func UserRelationShipToString(relation *UserRelationShip) string {
	data, _ := json.Marshal(relation)
	return string(data)
}

func NewUserRelation(userId, otherUserId int64, state string) *UserRelationShip {
	return &UserRelationShip{
		Id:        userId,
		State:     state,
		Otherside: otherUserId,
	}
}

func CheckRelationStateValid(state string) bool {
	if state == RELATION_STATE_DISLIKED ||
		state == RELATION_STATE_LIKED {
		return true
	}

	return false
}

func CheckUsrIdValid(id int64) bool {
	return id > 0
}
