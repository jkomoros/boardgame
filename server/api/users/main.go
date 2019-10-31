package users

//Factored into a sub-package so we don't get a cycle of dependencies between
//bolt and user db.

//StorageRecord denotes the storage record with info about a user.
type StorageRecord struct {
	//The Firebase user id
	ID          string
	Created     int64
	LastSeen    int64
	DisplayName string
	PhotoUrl    string
	Email       string
}

//EffectiveDisplayName returns a display name based on values in the
//StorageRecord.
func (s *StorageRecord) EffectiveDisplayName() string {
	if s.DisplayName != "" {
		return s.DisplayName
	}
	if s.Email != "" {
		return s.Email
	}
	return ""
}
