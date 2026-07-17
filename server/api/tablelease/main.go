// Package tablelease defines the persistence record for the single active
// shared Table display associated with a companion game.
package tablelease

import (
	"encoding/hex"
	"errors"
)

const (
	// TransitionTransfer marks a lease rotation performed through an explicit
	// Table-to-Table handoff. TransitionRecovery marks an expired-lease
	// recovery. Empty means that no user-facing transition reason is known.
	TransitionTransfer = "transfer"
	TransitionRecovery = "recovery"
	// TransitionSolo temporarily fences all transfer/recovery work while the
	// companion metadata is being irreversibly changed to solo mode.
	TransitionSolo = "solo"
)

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

	// A pending transfer has all four TransferID/digest/expiry fields and an
	// empty TransferTargetDeviceID. Redemption atomically rotates the lease and
	// fills TransferTargetDeviceID while retaining the remaining fields as a
	// short-lived receipt for lost-response retries. Secrets themselves are
	// never persisted.
	TransferID             string
	TransferTokenDigest    string
	TransferCodeDigest     string
	TransferExpires        int64
	TransferTargetDeviceID string

	// PreviousDeviceID and TransitionKind let a fenced-out display distinguish
	// an intentional handoff from ordinary lease recovery. They are metadata,
	// never authority.
	PreviousDeviceID string
	TransitionKind   string
}

// ValidateTransfer rejects partial or malformed transfer tuples before they
// reach durable storage. An empty tuple is valid and is how callers clear a
// pending transfer. This intentionally does not validate the older lease
// fields, whose storage contract historically allowed opaque test values.
func (r *StorageRecord) ValidateTransfer() error {
	if r == nil {
		return errors.New("nil companion Table lease")
	}
	if r.TransferID == "" && r.TransferTokenDigest == "" && r.TransferCodeDigest == "" &&
		r.TransferExpires == 0 && r.TransferTargetDeviceID == "" {
		return validateTransitionMetadata(r)
	}
	if !validHex(r.TransferID, 32) {
		return errors.New("Table transfer ID must be 32 lowercase hexadecimal characters")
	}
	if !validHex(r.TransferTokenDigest, 64) {
		return errors.New("Table transfer token digest must be 64 lowercase hexadecimal characters")
	}
	if !validHex(r.TransferCodeDigest, 64) {
		return errors.New("Table transfer code digest must be 64 lowercase hexadecimal characters")
	}
	if r.TransferExpires <= 0 {
		return errors.New("Table transfer expiry must be positive")
	}
	if r.TransferTargetDeviceID != "" && !validHex(r.TransferTargetDeviceID, 32) {
		return errors.New("Table transfer target device ID must be 32 lowercase hexadecimal characters")
	}
	if r.TransferTargetDeviceID == "" {
		if r.PreviousDeviceID != "" || r.TransitionKind != "" {
			return errors.New("pending Table transfer cannot contain transition metadata")
		}
	} else {
		if r.TransferTargetDeviceID != r.DeviceID {
			return errors.New("redeemed Table transfer target must be the active device")
		}
		if r.TransitionKind != TransitionTransfer || !validHex(r.PreviousDeviceID, 32) || r.PreviousDeviceID == r.DeviceID {
			return errors.New("redeemed Table transfer requires a distinct previous device and transfer transition")
		}
	}
	return validateTransitionMetadata(r)
}

func validateTransitionMetadata(r *StorageRecord) error {
	if r.PreviousDeviceID != "" && !validHex(r.PreviousDeviceID, 32) {
		return errors.New("previous Table device ID must be 32 lowercase hexadecimal characters")
	}
	if r.TransitionKind != "" && r.TransitionKind != TransitionTransfer && r.TransitionKind != TransitionRecovery && r.TransitionKind != TransitionSolo {
		return errors.New("invalid Table lease transition kind")
	}
	if r.TransitionKind != "" && r.PreviousDeviceID == "" {
		return errors.New("Table lease transition kind requires a previous device ID")
	}
	return nil
}

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded)*2 != length {
		return false
	}
	for _, char := range value {
		if char >= 'A' && char <= 'F' {
			return false
		}
	}
	return true
}

// TransferPending reports whether this record contains an unexpired transfer
// which has not yet selected a target. Invalid tuples are never considered
// usable, even if their timestamps happen to be in the future.
func (r *StorageRecord) TransferPending(nowMillis int64) bool {
	return r != nil && r.ValidateTransfer() == nil && r.TransferID != "" &&
		r.TransferTargetDeviceID == "" && r.TransferExpires > nowMillis
}

// TransferRedeemed reports whether this record contains the bounded receipt
// for a completed transfer. The receipt is useful only through its original
// transfer expiry.
func (r *StorageRecord) TransferRedeemed(nowMillis int64) bool {
	return r != nil && r.ValidateTransfer() == nil && r.TransferID != "" &&
		r.TransferTargetDeviceID != "" && r.TransferExpires > nowMillis
}

// ClearTransfer clears both pending handoff state and a redeemed retry receipt
// without disturbing the active lease or its transition metadata.
func (r *StorageRecord) ClearTransfer() {
	if r == nil {
		return
	}
	r.TransferID = ""
	r.TransferTokenDigest = ""
	r.TransferCodeDigest = ""
	r.TransferExpires = 0
	r.TransferTargetDeviceID = ""
}

// Clone returns a defensive copy suitable for handing across a storage API.
func (r *StorageRecord) Clone() *StorageRecord {
	if r == nil {
		return nil
	}
	result := *r
	return &result
}
