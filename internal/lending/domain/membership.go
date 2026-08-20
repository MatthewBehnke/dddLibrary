package domain

// MembershipStatus is a value object describing whether a Member may currently
// borrow. It is a small closed set, compared by value.
type MembershipStatus string

const (
	// MembershipActive members may borrow, subject to other rules.
	MembershipActive MembershipStatus = "active"
	// MembershipInactive members may not borrow.
	MembershipInactive MembershipStatus = "inactive"
)

// IsActive reports whether the status permits borrowing.
func (s MembershipStatus) IsActive() bool { return s == MembershipActive }

// Valid reports whether s is a recognised status.
func (s MembershipStatus) Valid() bool {
	switch s {
	case MembershipActive, MembershipInactive:
		return true
	default:
		return false
	}
}
