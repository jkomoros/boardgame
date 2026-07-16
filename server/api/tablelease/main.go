// Package tablelease defines the persistence record for the single active
// shared Table display associated with a companion game.
package tablelease

// StorageRecord is the durable state used to coordinate Table ownership
// across server processes. Generation is monotonically increased by storage
// on every successful compare-and-swap, including release; callers must not
// derive it themselves.
//
// SecretDigest stores a digest of the device capability, never the capability
// itself. HolderUserID and DeviceID are audit/display metadata and do not grant
// authority. Expires is a Unix millisecond timestamp.
type StorageRecord struct {
	GameID       string
	Generation   uint64
	DeviceID     string
	SecretDigest string
	HolderUserID string
	Expires      int64
}

// Clone returns a defensive copy suitable for handing across a storage API.
func (r *StorageRecord) Clone() *StorageRecord {
	if r == nil {
		return nil
	}
	result := *r
	return &result
}
