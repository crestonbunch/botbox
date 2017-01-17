package api

import (
	"time"
)

// A user contains all the relevant bits of information about a user contained
// in the database.
type User struct {
	Id            int       `json:"-" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"email" db:"email"`
	Joined        time.Time `json:"joined" db:"joined"`
	PermissionSet string    `json:"permission_set" db:"permission_set"`

	Permissions *PermissionSet `json:"-"`
	Profile     *Profile       `json:"profile"`
}

// A permission set is a set of permissions that a user has.
type PermissionSet struct {
	Name        string `db:"name"`
	Permissions []string
}

// A user's profile
type Profile struct {
	Id           int    `json:"-" `
	Bio          string `json:"bio" db:"bio"`
	Organization string `json:"organization" db:"organization"`
	Location     string `json:"location" db:"location"`
	Website      string `json:"website" db:"website"`
	Github       string `json:"github" db:"github"`
}

func (u *User) HasPermission(perm string) bool {
	return u.Permissions.HasPermission(perm)
}

func (p *PermissionSet) HasPermission(perm string) bool {
	for _, x := range p.Permissions {
		if x == perm {
			return true
		}
	}

	return false
}
