package api

import (
	"time"
)

// User is a struct containing information about a user.
type User struct {
	Id            int       `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Email         string    `json:"-" db:"email"`
	Joined        time.Time `json:"joined" db:"joined"`
	PermissionSet string    `json:"permission_set" db:"permission_set"`

	Permissions PermissionSet `json:"permissions"`
	Profile     *Profile      `json:"profile"`
}

// PermissionSet is a set of permissions that a user has.
type PermissionSet []string

// Profile is the public facing profile of a user.
type Profile struct {
	Id           int    `json:"-" `
	Bio          string `json:"bio" db:"bio"`
	Organization string `json:"organization" db:"organization"`
	Location     string `json:"location" db:"location"`
	Website      string `json:"website" db:"website"`
	Github       string `json:"github" db:"github"`
}

// HasPermission checks if a user has the given permission.
func (u *User) HasPermission(perm string) bool {
	return u.Permissions.HasPermission(perm)
}

// HasPermission checks if a permission set contains the given permission.
func (p PermissionSet) HasPermission(perm string) bool {
	for _, x := range p {
		if x == perm {
			return true
		}
	}

	return false
}
