package mysql

import (
	"github.com/jkomoros/boardgame/server/api/users"
)

//We define our own records, primarily to decorate with tags for Gorp, but
//also because e.g boardgame.storage.users.StorageRecord isn't structured the
//way we actually want to store in DB.

type UserStorageRecord struct {
	Id     string `db:",size:16"`
	Cookie string `db:",size:64"`
}

func (s *UserStorageRecord) ToStorageRecord() *users.StorageRecord {
	return &users.StorageRecord{
		Id: s.Id,
	}
}

func NewUserStorageRecord(user *users.StorageRecord) *UserStorageRecord {
	return &UserStorageRecord{
		Id: user.Id,
	}
}
