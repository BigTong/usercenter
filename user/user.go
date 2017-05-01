package user

import (
	"encoding/json"
)

const (
	RELATION_STATE_LIKED    = "liked"
	RELATION_STATE_DISLIKED = "disliked"
	RELATION_STATE_MATCHED  = "matched"
	RELATION_TYPE           = "relationship"
)

type User struct {
	_id         int64
	Id          int64  `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	Createdtime int64  `json:"createdTime,omitempty"`
	Address     string `json:"address,omitempty"`
	Type        string `json:"type,omitempty"`
}

type UserRelationShip struct {
	Id        int64  `json:"-"`
	State     string `json:"state,omitempty"`
	Otherside int64  `json:"user_id,string,omitempty"`
	Type      string `json:"type,omitempty"`
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
		Type:      RELATION_TYPE,
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
