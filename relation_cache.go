package usercenter

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"usercenter/user"

	"github.com/patrickmn/go-cache"
)

const (
	DEFAULT_BUCKET_NUM = 16
)

type UserRelations struct {
	rwLock        *sync.RWMutex
	relationShips []*user.UserRelationShip
}

func (u *UserRelations) UserRelations() string {
	u.rwLock.RLock()
	defer u.rwLock.Unlock()
	data, _ := json.Marshal(u.relationShips)
	return string(data)
}

func (u *UserRelations) LikeMe(otherUserId int64) bool {
	u.rwLock.RLock()
	defer u.rwLock.RUnlock()
	for _, r := range u.relationShips {
		if r.Otherside == otherUserId &&
			r.State == user.RELATION_STATE_LIKED {
			return true
		}
	}
	return false
}

func (u *UserRelations) UpdateUserRelation(
	userRelation *user.UserRelationShip) bool {
	u.rwLock.Lock()
	defer u.rwLock.Unlock()
	for _, r := range u.relationShips {
		if r.Otherside == userRelation.Otherside {
			r.State = userRelation.State
			return true
		}
	}

	u.relationShips = append(u.relationShips, userRelation)
	return true
}

func NewRelationsCache() *RelationsCache {
	ret := &RelationsCache{
		relationBuckets: []*cache.Cache{},
	}
	for i := 0; i < DEFAULT_BUCKET_NUM; i++ {
		ret.relationBuckets = append(ret.relationBuckets,
			cache.New(20*time.Minute, 10*time.Minute))
	}
	return ret

}

type RelationsCache struct {
	relationBuckets []*cache.Cache
}

func (r *RelationsCache) GetUserRelations(
	userId int64) (*UserRelations, bool) {
	bucketId := userId % 16
	val, ok := r.relationBuckets[bucketId].Get(fmt.Sprintf("%d", userId))
	if !ok {
		return nil, false
	}
	userRelations, _ := val.(*UserRelations)
	return userRelations, true
}

func (r *RelationsCache) SetUserRelations(userId int64,
	relationShips []*user.UserRelationShip) bool {
	bucketId := userId % 16

	val := &UserRelations{
		rwLock:        &sync.RWMutex{},
		relationShips: relationShips[0:],
	}

	r.relationBuckets[bucketId].Add(
		fmt.Sprintf("%d", userId), val, cache.DefaultExpiration)
	return true
}
