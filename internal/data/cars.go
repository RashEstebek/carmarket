package data

import (
	"carMarket.dreamteam.kz/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Car struct {
	ID          int64  `json:"id"`
	Model       string `json:"model"`
	Year        int64  `json:"year"`
	Price       int64  `json:"price"`
	Marka       string `json:"marka"`
	Color       string `json:"color"`
	Type        string `json:"type,omitempty"`
	Image       string `json:"image,omitempty"`
	Description string `json:"description,omitempty"`
}

func ValidateCar(v *validator.Validator, car *Car) {
	v.Check(car.Model != "", "model", "must be provided")
	v.Check(car.Marka != "", "marka", "must be provided")
	v.Check(car.Color != "", "color", "must be provided")
	v.Check(car.Type != "", "type", "must be provided")
	v.Check(car.Image != "", "image", "must be provided")
	v.Check(len(car.Model) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(car.Price != 0, "price", "must be provided")
	v.Check(car.Price > 0, "price", "must be a positive integer")
	v.Check(car.Year > 1800, "year", "must be greater than 1800")
	v.Check(car.Year != 0, "year", "must be provided")
	v.Check(car.Description != "", "description", "must be provided")

}

type CarModel struct {
	DB *sql.DB
}

func (c CarModel) Insert(car *Car) error {
	query := `INSERT INTO cars (model,year, price, marka, color, type, image, description)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				RETURNING id`
	args := []any{car.Model, car.Year, car.Price, car.Marka, car.Color, car.Type, car.Image, car.Description}
	return c.DB.QueryRow(query, args...).Scan(&car.ID)
}

func (c CarModel) Get(id int64) (*Car, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT *
		FROM cars
		WHERE id = $1`
	var car Car

	err := c.DB.QueryRow(query, id).Scan(
		&car.ID,
		&car.Model,
		&car.Year,
		&car.Price,
		&car.Marka,
		&car.Color,
		&car.Type,
		&car.Image,
		&car.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &car, nil
}

func (c CarModel) Update(car *Car) error {
	query := `
			UPDATE cars
			SET model = $1, year=$2, price = $3, marka = $4, color = $5, type = $6, 
			 image = $7, description=$8 
			WHERE id = $9
			RETURNING id`
	args := []any{
		car.Model,
		car.Year,
		car.Price,
		car.Marka,
		car.Color,
		car.Type,
		car.Image,
		car.Description,
		car.ID,
	}
	return c.DB.QueryRow(query, args...).Scan(&car.ID)
}

func (c CarModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
				DELETE FROM cars
				WHERE id = $1`
	result, err := c.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (c CarModel) GetAll(model string, year int64, marka string, priceMax int64, priceMin int64,
	color string, type_ string, description string,
	filters Filters) ([]*Car, error) {

	query := fmt.Sprintf(`
								SELECT *
								FROM cars
								WHERE (to_tsvector('simple', model) @@ plainto_tsquery('simple', $1) OR $1 = '')
								AND price < $2 AND price > $3
								ORDER BY %s %s, id ASC LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, model, priceMax, priceMin, filters.limit(),
		filters.offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cars := []*Car{}
	for rows.Next() {
		var car Car

		err := rows.Scan(
			&car.ID,
			&car.Model,
			&car.Year,
			&car.Price,
			&car.Marka,
			&car.Color,
			&car.Type,
			&car.Image,
			&car.Description,
		)
		if err != nil {
			return nil, err
		}
		cars = append(cars, &car)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cars, nil
}
