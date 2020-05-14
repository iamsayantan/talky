package store

import "github.com/iamsayantan/talky"

// UserRepository provides the interface for the user storage.
type UserRepository interface {
	CreateUser(user *talky.User) (*talky.User, error)
	FindById(id uint) (*talky.User, error)
	FindByUsername(username string) (*talky.User, error)
}
