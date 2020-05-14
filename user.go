package talky

import (
	"errors"
	"time"
)

type User struct {
	ID        uint       `gorm:"primary_key" json:"id"`
	Username  string     `json:"username"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Password  string     `json:"-"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

func (u *User) IsValid() error {
	var err error

	if u.FirstName == "" {
		err = errors.New("first name can not be left blank")
		return err
	}

	if u.LastName == "" {
		err = errors.New("last name can not be left blank")
		return err
	}

	if u.Username == "" {
		err = errors.New("username can not be left blank")
		return err
	}

	if u.Password == "" {
		err = errors.New("password can not be left blank")
		return err
	}

	return nil
}
