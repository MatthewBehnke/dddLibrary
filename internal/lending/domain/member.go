package domain

// MaxActiveLoans is the maximum number of concurrent active loans a Member may
// hold.
const MaxActiveLoans = 5

// Member is an aggregate root representing a library patron. Its borrowing
// eligibility depends on membership status and current active-loan count.
type Member struct {
	id          MemberID
	name        string
	status      MembershipStatus
	activeLoans int
}

// NewMember creates a fresh active Member with no loans and a domain-minted
// identity.
func NewMember(name string) *Member {
	return &Member{id: NewMemberID(), name: name, status: MembershipActive}
}

// ReconstituteMember rebuilds a Member from persisted state. Repositories use
// this; application code should not.
func ReconstituteMember(id MemberID, name string, status MembershipStatus, activeLoans int) *Member {
	return &Member{id: id, name: name, status: status, activeLoans: activeLoans}
}

func (m *Member) ID() MemberID             { return m.id }
func (m *Member) Name() string             { return m.name }
func (m *Member) Status() MembershipStatus { return m.status }
func (m *Member) ActiveLoans() int         { return m.activeLoans }

// CanBorrow reports whether the Member is eligible to take another loan,
// returning a typed domain error describing the first violated rule.
func (m *Member) CanBorrow() error {
	if !m.status.IsActive() {
		return ErrMembershipInactive
	}
	if m.activeLoans >= MaxActiveLoans {
		return ErrLoanLimitReached
	}
	return nil
}

// recordBorrow increments the active-loan count. Unexported: only the domain
// (the Borrow service) may mutate it.
func (m *Member) recordBorrow() { m.activeLoans++ }
