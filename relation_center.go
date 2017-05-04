package usercenter

import (
	"encoding/json"
	"time"

	"usercenter/db"
	"usercenter/user"

	"github.com/BigTong/common/log"
)

const (
	DEFAULT_CHAN_LEN = 100
)

func NewRelationShipCenter() *RelationShipCenter {
	ret := &RelationShipCenter{
		relationsToDb:         make(chan *user.UserRelationShip, DEFAULT_CHAN_LEN),
		cache:                 NewRelationsCache(),
		postgresDb:            db.NewPostgresQlDb(*postgresdbConfigFile),
		needFlushRelationData: false,
	}
	go ret.writeUserRelationsToDb()
	return ret
}

type RelationShipCenter struct {
	relationsToDb         chan *user.UserRelationShip
	cache                 *RelationsCache
	postgresDb            db.UserDao
	needFlushRelationData bool
}

func (rls *RelationShipCenter) GetUserRelationShip(userId int64) string {
	relations, ok := rls.cache.GetUserRelations(userId)
	if ok {
		return relations.UserRelations()
	}
	relationsArray, err := rls.postgresDb.GetUserRelation(userId)
	if err != nil {
		log.FInfo("read user relations get err: %s", err.Error())
		return "[]"
	}

	rls.cache.SetUserRelations(userId, relationsArray[0:])

	data, _ := json.Marshal(relationsArray)
	return string(data)
}

func (rls *RelationShipCenter) UpdateRelationShip(relation *user.UserRelationShip) string {
	userRelations, ok := rls.cache.GetUserRelations(relation.Id)
	if relation.State == user.RELATION_STATE_DISLIKED {
		rls.relationsToDb <- relation
		if ok {
			userRelations.UpdateUserRelation(relation)
		}
		return user.UserRelationShipToString(relation)
	}

	otherUserRelations, otherOk := rls.cache.GetUserRelations(relation.Otherside)
	if otherOk && otherUserRelations.LikeMe(relation.Id) {
		rls.relationsToDb <- relation
		relation.State = user.RELATION_STATE_MATCHED
		if ok {
			userRelations.UpdateUserRelation(relation)
		}
		ret := user.UserRelationShipToString(relation)
		relation.Id, relation.Otherside = relation.Otherside, relation.Id
		otherUserRelations.UpdateUserRelation(relation)
		return ret
	}

	newRelation, err := rls.postgresDb.UpdateUserRelation(relation)
	if err != nil {
		log.FFatal("update postgres db get err:%s", err.Error())
	}
	if ok {
		userRelations.UpdateUserRelation(newRelation)
	}

	return user.UserRelationShipToString(newRelation)
}

func (rls *RelationShipCenter) waitingForDataWriteFinished() bool {
	rls.needFlushRelationData = true
	for {
		if len(rls.relationsToDb) == 0 {
			rls.postgresDb.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return true
}

func (rls *RelationShipCenter) writeUserRelationsToDb() {
	cnt := 0
	relations := []*user.UserRelationShip{}
	for {
		relation := <-rls.relationsToDb
		relations = append(relations, relation)
		cnt++
		if (rls.needFlushRelationData && len(rls.relationsToDb) == 0) ||
			cnt == DEFAULT_BATCH_WRITE_NUM {
			err := rls.postgresDb.UpdateUserRelations(relations)
			if err != nil {
				log.FFatal("write db get error:%s", err.Error())
			}
			cnt = 0
			relations = relations[:0]
		}

	}
}
