package models

import "testing"

func TestMemberNoValidateTypes(t *testing.T) {
	if err := (Member{}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (MemberList{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
