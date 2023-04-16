package data

import (
	"context"
	"database/sql"
	"time"
)

type CartsModel struct {
	DB *sql.DB
}

func (c CartsModel) AddToCart(userID int64, car Car) error {
	query := `UPDATE carts
	SET cars_id = array_append(cars_id, $1)
	WHERE user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.DB.ExecContext(ctx, query, car.ID, userID)
	return err
}

func (c CartsModel) CreateCart(userID int64) error {
	query := `
INSERT INTO carts VALUES ($1)`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := c.DB.ExecContext(ctx, query, userID)
	return err
}
