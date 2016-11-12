package api

import (
	"errors"
)

type Expectation interface {
	fulfilled() bool
	String() string
}

type mockUserModel struct {
	Expectations []Expectation
	Order        []int
}

func newMockUserModel() *mockUserModel {
	return &mockUserModel{Expectations: []Expectation{}, Order: []int{}}
}

func (m *mockUserModel) ExpectationsWereMetInOrder() error {
	for _, e := range m.Expectations {
		if !e.fulfilled() {
			return errors.New(e.String() + " was not fulfilled!")
		}
	}

	if len(m.Expectations) != len(m.Order) {
		return errors.New("User model did not meet all expectations!")
	}

	for i := 1; i < len(m.Order); i++ {
		if m.Order[i-1] > m.Order[i] {
			return errors.New("Expectations were met in the wrong order!")
		}
	}

	return nil
}

func (m *mockUserModel) SelectByUsername(username string) (*User, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *SelectByUsernameExpectation:
			casted := e.(*SelectByUsernameExpectation)
			if casted.username == username && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.user, casted.err
			}
		}
	}
	return nil, errors.New("Call to SelectByUsername was unexpected.")
}

func (m *mockUserModel) ExpectSelectByUsername() *SelectByUsernameExpectation {
	e := &SelectByUsernameExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type SelectByUsernameExpectation struct {
	username string
	err      error
	user     *User
	done     bool
}

func (e *SelectByUsernameExpectation) String() string {
	return "SelectByUsername"
}

func (e *SelectByUsernameExpectation) WithArgs(
	username string,
) *SelectByUsernameExpectation {
	e.username = username
	return e
}

func (e *SelectByUsernameExpectation) WillReturn(
	user *User,
) *SelectByUsernameExpectation {
	e.user = user
	return e
}

func (e *SelectByUsernameExpectation) WillReturnError(
	err error,
) *SelectByUsernameExpectation {
	e.err = err
	return e
}

func (e *SelectByUsernameExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) SelectByEmail(email string) (*User, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *SelectByEmailExpectation:
			casted := e.(*SelectByEmailExpectation)
			if casted.email == email && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.user, casted.err
			}
		}
	}
	return nil, errors.New("Call to SelectByEmail was unexpected.")
}

func (m *mockUserModel) ExpectSelectByEmail() *SelectByEmailExpectation {
	e := &SelectByEmailExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type SelectByEmailExpectation struct {
	email string
	err   error
	user  *User
	done  bool
}

func (e *SelectByEmailExpectation) String() string {
	return "SelectByEmail"
}

func (e *SelectByEmailExpectation) WithArgs(
	email string,
) *SelectByEmailExpectation {
	e.email = email
	return e
}

func (e *SelectByEmailExpectation) WillReturn(
	user *User,
) *SelectByEmailExpectation {
	e.user = user
	return e
}

func (e *SelectByEmailExpectation) WillReturnError(
	err error,
) *SelectByEmailExpectation {
	e.err = err
	return e
}

func (e *SelectByEmailExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) SelectByUsernameAndEmail(
	username, email string,
) (*User, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *SelectByUsernameAndEmailExpectation:
			casted := e.(*SelectByUsernameAndEmailExpectation)
			if casted.user.Username == username && casted.email == email &&
				!casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.user, casted.err
			}
		}
	}
	return nil, errors.New("Call to SelectByUsernameAndEmail was unexpected.")
}

func (
	m *mockUserModel,
) ExpectSelectByUsernameAndEmail() *SelectByUsernameAndEmailExpectation {
	e := &SelectByUsernameAndEmailExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type SelectByUsernameAndEmailExpectation struct {
	username string
	email    string
	err      error
	user     *User
	done     bool
}

func (e *SelectByUsernameAndEmailExpectation) String() string {
	return "SelectByUsernameAndEmail"
}

func (e *SelectByUsernameAndEmailExpectation) WithArgs(
	username string,
	email string,
) *SelectByUsernameAndEmailExpectation {
	e.username = username
	e.email = email
	return e
}

func (e *SelectByUsernameAndEmailExpectation) WillReturn(
	user *User,
) *SelectByUsernameAndEmailExpectation {
	e.user = user
	return e
}

func (e *SelectByUsernameAndEmailExpectation) WillReturnError(
	err error,
) *SelectByUsernameAndEmailExpectation {
	e.err = err
	return e
}

func (e *SelectByUsernameAndEmailExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) SelectBySession(session string) (*User, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *SelectBySessionExpectation:
			casted := e.(*SelectBySessionExpectation)
			if casted.session == session && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.user, casted.err
			}
		}
	}
	return nil, errors.New("Call to SelectBySession was unexpected.")
}

func (m *mockUserModel) ExpectSelectBySession() *SelectBySessionExpectation {
	e := &SelectBySessionExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type SelectBySessionExpectation struct {
	session string
	err     error
	user    *User
	done    bool
}

func (e *SelectBySessionExpectation) String() string {
	return "SelectBySessionExpectation"
}

func (e *SelectBySessionExpectation) WithArgs(
	session string,
) *SelectBySessionExpectation {
	e.session = session
	return e
}

func (e *SelectBySessionExpectation) WillReturn(
	user *User,
) *SelectBySessionExpectation {
	e.user = user
	return e
}

func (e *SelectBySessionExpectation) WillReturnError(
	err error,
) *SelectBySessionExpectation {
	e.err = err
	return e
}

func (e *SelectBySessionExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) Insert(
	username, email, password string,
) (int, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *InsertExpectation:
			casted := e.(*InsertExpectation)
			if casted.username == username && casted.email == email &&
				casted.password == password && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.result, casted.err
			}
		}
	}
	return 0, errors.New("Call to Insert was unexpected.")
}

func (m *mockUserModel) ExpectInsert() *InsertExpectation {
	e := &InsertExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type InsertExpectation struct {
	username string
	email    string
	password string
	err      error
	result   int
	done     bool
}

func (e *InsertExpectation) String() string {
	return "Insert"
}

func (e *InsertExpectation) WithArgs(
	username, email, password string,
) *InsertExpectation {
	e.username = username
	e.email = email
	e.password = password
	return e
}

func (e *InsertExpectation) WillReturn(
	i int,
) *InsertExpectation {
	e.result = i
	return e
}

func (e *InsertExpectation) WillReturnError(
	err error,
) *InsertExpectation {
	e.err = err
	return e
}

func (e *InsertExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) Update(
	user *User, fullname, email, bio, organization, location string,
) error {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *UpdateExpectation:
			casted := e.(*UpdateExpectation)
			if casted.user == user && casted.fullname == fullname &&
				casted.email == email && casted.bio == bio && casted.organization ==
				organization && casted.location == location && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.err
			}
		}
	}
	return errors.New("Call to Update was unexpected.")
}

func (m *mockUserModel) ExpectUpdate() *UpdateExpectation {
	e := &UpdateExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type UpdateExpectation struct {
	user         *User
	fullname     string
	email        string
	bio          string
	organization string
	location     string
	err          error
	result       int
	done         bool
}

func (e *UpdateExpectation) String() string {
	return "Update"
}

func (e *UpdateExpectation) WithArgs(
	user *User, fullname, email, bio, organization, location string,
) *UpdateExpectation {
	e.user = user
	e.fullname = fullname
	e.email = email
	e.bio = bio
	e.organization = organization
	e.location = location
	return e
}

func (e *UpdateExpectation) WillReturnError(
	err error,
) *UpdateExpectation {
	e.err = err
	return e
}

func (e *UpdateExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) VerifyPassword(
	username, password string,
) (bool, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *VerifyPasswordExpectation:
			casted := e.(*VerifyPasswordExpectation)
			if casted.username == username && casted.password == password &&
				casted.done != true {
				m.Order = append(m.Order, i)
				return casted.result, casted.err
			}
		}
	}
	return false, errors.New("Call to VerifyPassword was unexpected.")
}

func (m *mockUserModel) ExpectVerifyPassword() *VerifyPasswordExpectation {
	e := &VerifyPasswordExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type VerifyPasswordExpectation struct {
	username string
	password string
	err      error
	result   bool
	done     bool
}

func (e *VerifyPasswordExpectation) String() string {
	return "VerifyPassword"
}

func (e *VerifyPasswordExpectation) WithArgs(
	username, password string,
) *VerifyPasswordExpectation {
	e.username = username
	e.password = password
	return e
}

func (e *VerifyPasswordExpectation) WillReturn(
	result bool,
) *VerifyPasswordExpectation {
	e.result = result
	return e
}

func (e *VerifyPasswordExpectation) WillReturnError(
	err error,
) *VerifyPasswordExpectation {
	e.err = err
	return e
}

func (e *VerifyPasswordExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) ChangePassword(
	user *User, password string,
) error {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *ChangePasswordExpectation:
			casted := e.(*ChangePasswordExpectation)
			if casted.user == user && casted.password == password &&
				casted.done != true {
				m.Order = append(m.Order, i)
				return casted.err
			}
		}
	}
	return errors.New("Call to ChangePassword was unexpected.")
}

func (m *mockUserModel) ExpectChangePassword() *ChangePasswordExpectation {
	e := &ChangePasswordExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type ChangePasswordExpectation struct {
	user     *User
	password string
	err      error
	result   bool
	done     bool
}

func (e *ChangePasswordExpectation) String() string {
	return "ChangePassword"
}

func (e *ChangePasswordExpectation) WithArgs(
	user *User, password string,
) *ChangePasswordExpectation {
	e.user = user
	e.password = password
	return e
}

func (e *ChangePasswordExpectation) WillReturnError(
	err error,
) *ChangePasswordExpectation {
	e.err = err
	return e
}

func (e *ChangePasswordExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) ChangePermissions(
	user *User, permissions string,
) error {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *ChangePermissionsExpectation:
			casted := e.(*ChangePermissionsExpectation)
			if casted.user == user && casted.permissions == permissions &&
				casted.done != true {
				m.Order = append(m.Order, i)
				return casted.err
			}
		}
	}
	return errors.New("Call to ChangePermissions was unexpected.")
}

func (
	m *mockUserModel,
) ExpectChangePermissions() *ChangePermissionsExpectation {
	e := &ChangePermissionsExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type ChangePermissionsExpectation struct {
	user        *User
	permissions string
	err         error
	result      bool
	done        bool
}

func (e *ChangePermissionsExpectation) String() string {
	return "ChangePassword"
}

func (e *ChangePermissionsExpectation) WithArgs(
	user *User, perm string,
) *ChangePermissionsExpectation {
	e.user = user
	e.permissions = perm
	return e
}

func (e *ChangePermissionsExpectation) WillReturnError(
	err error,
) *ChangePermissionsExpectation {
	e.err = err
	return e
}

func (e *ChangePermissionsExpectation) fulfilled() bool {
	return e.done
}

func (m *mockUserModel) CreateVerificationSecret(
	id int, email string,
) (string, error) {
	for i, e := range m.Expectations {
		switch e.(type) {
		case *CreateVerificationSecretExpectation:
			casted := e.(*CreateVerificationSecretExpectation)
			if casted.id == id && casted.email == email && !casted.done {
				casted.done = true
				m.Order = append(m.Order, i)
				return casted.secret, casted.err
			}
		}
	}
	return "", errors.New("Call to CreateVerificationSecret was unexpected.")
}

func (
	m *mockUserModel,
) ExpectCreateVerificationSecret() *CreateVerificationSecretExpectation {
	e := &CreateVerificationSecretExpectation{}
	m.Expectations = append(m.Expectations, e)
	return e
}

type CreateVerificationSecretExpectation struct {
	id     int
	email  string
	err    error
	secret string
	done   bool
}

func (e *CreateVerificationSecretExpectation) String() string {
	return "CreateVerification"
}

func (e *CreateVerificationSecretExpectation) WithArgs(
	id int, email string,
) *CreateVerificationSecretExpectation {
	e.id = id
	e.email = email
	return e
}

func (e *CreateVerificationSecretExpectation) WillReturn(
	secret string,
) *CreateVerificationSecretExpectation {
	e.secret = secret
	return e
}

func (e *CreateVerificationSecretExpectation) WillReturnError(
	err error,
) *CreateVerificationSecretExpectation {
	e.err = err
	return e
}

func (e *CreateVerificationSecretExpectation) fulfilled() bool {
	return e.done
}
