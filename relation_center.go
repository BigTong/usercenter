package usercenter

import (
	"encoding/json"
	"log"

	"usercenter/db"
	"usercenter/user"
)

const (
	DEFAULT_CHAN_LEN = 100
)

func NewRelationShipCenter() *RelationShipCenter {
	ret := &RelationShipCenter{}
	go ret.writeUserRelationsToDb()
	return ret
}

type RelationShipCenter struct {
	relationsToDb         chan *user.UserRelationShip
	cache                 *RelationsCache
	postgresDb            db.UserDao
	needFlushRelationData bool
}

func (self *RelationShipCenter) GetUserRelationShip(userId int64) string {
	relations, ok := self.cache.GetUserRelations(userId)
	if ok {
		return relations.UserRelations()
	}
	relationsArray, err := self.postgresDb.GetUserRelation(userId)
	if err != nil {
		log.Printf("read user relations get err: %s", err.Error())
		return "[]"
	}

	self.cache.SetUserRelations(userId, relationsArray[0:])

	data, _ := json.Marshal(relationsArray)
	return string(data)
}

func (self *RelationShipCenter) UpdateRelationShip(relation *user.UserRelationShip) string {
	userRelations, ok := self.cache.GetUserRelations(relation.Id)
	if relation.State == user.RELATION_STATE_DISLIKED {
		self.relationsToDb <- relation
		if ok {
			userRelations.UpdateUserRelation(relation)
		}
		return user.UserRelationShipToString(relation)
	}

	otherUserRelations, otherOk := self.cache.GetUserRelations(relation.OtherSide)
	if otherOk && otherUserRelations.LikeMe(relation.Id) {
		self.relationsToDb <- relation
		relation.State = user.RELATION_STATE_MATCHED
		if ok {
			userRelations.UpdateUserRelation(relation)
		}
		ret := user.UserRelationShipToString(relation)
		relation.Id, relation.OtherSide = relation.OtherSide, relation.Id
		otherUserRelations.UpdateUserRelation(relation)
		return ret
	}

	newRelation, err := self.postgresDb.UpdateUserRelation(relation)
	if err != nil {
		log.Panic("update postgres db get err:" + err.Error())
	}
	if ok {
		userRelations.UpdateUserRelation(newRelation)
	}

	return user.UserRelationShipToString(newRelation)
}

func (self *RelationShipCenter) writeUserRelationsToDb() {
	cnt := 0
	relations := make([]*user.UserRelationShip, DEFAULT_BATCH_WRITE_NUM)
	for {
		relation := <-self.relationsToDb
		relations = append(relations, relation)
		cnt++
		if (self.needFlushRelationData && len(self.relationsToDb) == 0) ||
			cnt == DEFAULT_BATCH_WRITE_NUM {
			err := self.postgresDb.UpdateUserRelations(relations)
			if err != nil {
				log.Panic("write db get error:" + err.Error())
			}
			cnt = 0
			relations = relations[:0]
		}

	}
}
