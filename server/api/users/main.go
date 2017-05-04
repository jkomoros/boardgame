package users

//Factored into a sub-package so we don't get a cycle of dependencies between
//bolt and user db.

type StorageRecord struct {
	//The Firebase user id
	Id          string
	Created     int64
	LastSeen    int64
	DisplayName string
	PhotoUrl    string
	Email       string
}

func (s *StorageRecord) EffectiveDisplayName() string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	if s.Email != "" {
		return s.Email
	}
	return ""
}
