package api

import (
	"testing"
)

func TestUserHasPermission(t *testing.T) {
	u := &User{
		Permissions: PermissionSet([]string{
			"POST_COMMENT", "UPLOAD_FILE", "EDIT_PROFILE",
		}),
	}

	testCases := []struct {
		SearchFor string
		Expect    bool
	}{
		{SearchFor: "POST_COMMENT", Expect: true},
		{SearchFor: "UPLOAD_FILE", Expect: true},
		{SearchFor: "EDIT_PROFILE", Expect: true},
		{SearchFor: "NOT_A_PERMISSION", Expect: false},
	}

	for _, test := range testCases {
		result := u.HasPermission(test.SearchFor)

		if result != test.Expect {
			t.Error("Check permissions was not correct.")
		}
	}
}
