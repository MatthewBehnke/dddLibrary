package domain

import "testing"

func TestMembershipStatus_IsActive(t *testing.T) {
	if !MembershipActive.IsActive() {
		t.Error("active membership should report IsActive() true")
	}
	if MembershipInactive.IsActive() {
		t.Error("inactive membership should report IsActive() false")
	}
}

func TestMembershipStatus_Valid(t *testing.T) {
	if !MembershipActive.Valid() || !MembershipInactive.Valid() {
		t.Error("known statuses should be valid")
	}
	if MembershipStatus("bogus").Valid() {
		t.Error("unknown status should be invalid")
	}
}
