package api

import (
	"time"
)

// A user contains all the relevant bits of information about a user contained
// in the database.
type User struct {
	Id            int       `json:"-"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	Joined        time.Time `json:"joined"`
	PermissionSet string    `json:"permission_set"`

	Permissions *PermissionSet `json:"-"`
	Profile     *Profile       `json:"profile"`
}

func (u *User) HasPermission(perm string) bool {
	return u.Permissions.HasPermission(perm)
}

// A permission set is a set of permissions that a user has.
type PermissionSet struct {
	Name        string
	Permissions []string
}

func (p *PermissionSet) HasPermission(perm string) bool {
	for _, x := range p.Permissions {
		if x == perm {
			return true
		}
	}

	return false
}

// A user's profile
type Profile struct {
	Id           int    `json:"-"`
	Bio          string `json:"bio"`
	Organization string `json:"organization"`
	Location     string `json:"location"`
	Website      string `json:"website"`
	Github       string `json:"github"`
}
