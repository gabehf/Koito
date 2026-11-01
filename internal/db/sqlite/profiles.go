package sqlite

import (
	"context"
	"fmt"
)

// resolveWriteProfile keeps legacy UserID callers working during step 1.
// An explicit profile must exist; an unknown legacy user never falls back to default.
func (s *Sqlite) resolveWriteProfile(ctx context.Context, profileID, userID int32) (int32, error) {
	if profileID > 0 {
		var id int32
		if err := s.db.QueryRowContext(ctx, `SELECT id FROM listening_profiles WHERE id = ?`, profileID).Scan(&id); err != nil {
			return 0, fmt.Errorf("profile not found: %w", err)
		}
		if userID == 0 {
			return id, nil
		}
	}
	if userID <= 0 || profileID < 0 {
		return 0, fmt.Errorf("a valid ProfileID or legacy UserID is required")
	}
	var legacyProfileID int32
	if err := s.db.QueryRowContext(ctx, `SELECT profile_id FROM legacy_user_profiles WHERE user_id = ?`, userID).Scan(&legacyProfileID); err != nil {
		return 0, fmt.Errorf("legacy user has no listening profile: %w", err)
	}
	if profileID != 0 && profileID != legacyProfileID {
		return 0, fmt.Errorf("ProfileID does not match legacy UserID")
	}
	return legacyProfileID, nil
}

func nullableLegacyUserID(id int32) any {
	if id == 0 {
		return nil
	}
	return id
}
