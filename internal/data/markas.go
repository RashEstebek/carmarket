package data

import (
	"carMarket.dreamteam.kz/internal/validator"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Marka struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Producer string `json:"producer"`
	Logo     string `json:"logo,omitempty"`
}

func ValidateMarka(v *validator.Validator, marka *Marka) {
	v.Check(marka.Name != "", "name", "must be provided")
	v.Check(marka.Producer != "", "producer", "must be provided")
	v.Check(marka.Logo != "", "logo", "must be provided")
}

type MarkaModel struct {
	DB *sql.DB
}

func (m MarkaModel) Insert(marka *Marka) error {
	query := `INSERT INTO markas (name,producer,logo)
				VALUES ($1, $2, $3)
				RETURNING id`
	args := []any{marka.Name, marka.Producer, marka.Logo}
	return m.DB.QueryRow(query, args...).Scan(&marka.ID)
}

func (m MarkaModel) Get(id int64) (*Marka, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT *
		FROM markas
		WHERE id = $1`
	var marka Marka

	err := m.DB.QueryRow(query, id).Scan(
		&marka.ID,
		&marka.Name,
		&marka.Producer,
		&marka.Logo,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &marka, nil
}

func (m MarkaModel) Update(marka *Marka) error {
	query := `
			UPDATE markas
			SET name = $1, country = $2, logo = $3
			WHERE id = $4
			RETURNING id`
	args := []any{
		marka.Name,
		marka.Producer,
		marka.Logo,
		marka.ID,
	}
	return m.DB.QueryRow(query, args...).Scan(&marka.ID)
}

func (m MarkaModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
				DELETE FROM markas
				WHERE id = $1`
	result, err := m.DB.Exec(query, id)
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

func (m MarkaModel) GetAll() ([]*Marka, error) {
	query := fmt.Sprintf(`
								SELECT *
								FROM markas ORDER BY id`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	markas := []*Marka{}
	for rows.Next() {
		var marka Marka

		err := rows.Scan(
			&marka.ID,
			&marka.Name,
			&marka.Producer,
			&marka.Logo,
		)
		if err != nil {
			return nil, err
		}
		markas = append(markas, &marka)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return markas, nil
}
