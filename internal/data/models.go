package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Cars   CarModel
	Users  UserModel
	Markas MarkaModel
	Tokens TokenModel
	Carts  CartsModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Cars:   CarModel{DB: db},
		Users:  UserModel{DB: db},
		Markas: MarkaModel{DB: db},
		Tokens: TokenModel{DB: db},
		Carts:  CartsModel{DB: db},
	}
}
