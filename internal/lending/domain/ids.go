package domain

import "github.com/google/uuid"

// BookID, MemberID, and LoanID are typed identity wrappers around uuid.UUID.
// Typing them separately makes it a compile error to pass a MemberID where a
// BookID is expected, and keeps identity a domain concern rather than a
// persistence detail.

type (
	// BookID identifies a Book aggregate.
	BookID uuid.UUID
	// MemberID identifies a Member aggregate.
	MemberID uuid.UUID
	// LoanID identifies a Loan aggregate.
	LoanID uuid.UUID
)

// NewBookID mints a fresh random BookID.
func NewBookID() BookID { return BookID(uuid.New()) }

// NewMemberID mints a fresh random MemberID.
func NewMemberID() MemberID { return MemberID(uuid.New()) }

// NewLoanID mints a fresh random LoanID.
func NewLoanID() LoanID { return LoanID(uuid.New()) }

func (id BookID) String() string   { return uuid.UUID(id).String() }
func (id MemberID) String() string { return uuid.UUID(id).String() }
func (id LoanID) String() string   { return uuid.UUID(id).String() }

// ParseBookID parses a canonical UUID string into a BookID.
func ParseBookID(s string) (BookID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return BookID{}, err
	}
	return BookID(u), nil
}

// ParseMemberID parses a canonical UUID string into a MemberID.
func ParseMemberID(s string) (MemberID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return MemberID{}, err
	}
	return MemberID(u), nil
}

// ParseLoanID parses a canonical UUID string into a LoanID.
func ParseLoanID(s string) (LoanID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return LoanID{}, err
	}
	return LoanID(u), nil
}
