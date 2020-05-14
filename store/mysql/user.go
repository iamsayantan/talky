package mysql

import (
	"errors"
	"github.com/iamsayantan/talky"
	"github.com/iamsayantan/talky/store"
	"github.com/jinzhu/gorm"
)

var (
	ErrInvalidUserDetails = errors.New("invalid user details")
)

type userRepository struct {
	db *gorm.DB
}

func (ur *userRepository) CreateUser(user *talky.User) (*talky.User, error) {
	if user.ID != 0 {
		return nil, ErrInvalidUserDetails
	}

	err := ur.db.Create(&user).Error
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *userRepository) FindById(id uint) (*talky.User, error) {
	user := &talky.User{}
	if err := ur.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (ur *userRepository) FindByUsername(username string) (*talky.User, error) {
	user := &talky.User{}
	if err := ur.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func NewUserRepository(db *gorm.DB) store.UserRepository {
	return &userRepository{db: db}
}
