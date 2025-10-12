// path: backend/internal/domain/user/types.go
// CREATE THIS NEW FILE or ADD TO user.go

package user

import (
	"time"

	"github.com/google/uuid"
)

// Statistics holds user statistics
type Statistics struct {
	TotalUsers       int64 `json:"totalUsers"`
	ActiveUsers      int64 `json:"activeUsers"`
	VerifiedUsers    int64 `json:"verifiedUsers"`
	NewUsersThisWeek int64 `json:"newUsersThisWeek"`
}

// ADD THESE TO YOUR backend/internal/domain/user/user.go FILE:

// SetID sets the user ID (used by repository after creation)
func (u *User) SetID(id uuid.UUID) {
	u.id = id
}

// PasswordHash returns the password hash
func (u *User) PasswordHash() string {
	return u.passwordHash
}

func (u *User) RecordLogin(ipAddress string) error {
	if u.status == StatusSuspended {
		return ErrAccountSuspended
	}

	now := time.Now().UTC()
	u.lastLoginAt = &now
	u.updatedAt = now

	// You can store IP address if you have a field for it
	// u.lastLoginIP = ipAddress

	return nil
}
