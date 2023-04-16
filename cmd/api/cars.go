package main

import (
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/validator"
	"errors"
	"fmt"
	"net/http"
)

func (app *application) createCarHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Model       string `json:"model"`
		Year        int64  `json:"year"`
		Price       int64  `json:"price"`
		Marka       string `json:"marka"`
		Color       string `json:"color"`
		Type        string `json:"type"`
		Image       string `json:"image"`
		Description string `json:"description"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	car := &data.Car{
		Model:       input.Model,
		Year:        input.Year,
		Price:       input.Price,
		Marka:       input.Marka,
		Color:       input.Color,
		Type:        input.Type,
		Image:       input.Image,
		Description: input.Description,
	}
	v := validator.NewToActivate()

	if data.ValidateCar(v, car); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Cars.Insert(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/cars/%d", car.ID))
	err = app.writeJSON(w, http.StatusCreated, car, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	car, err := app.models.Cars.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Encode the struct to JSON and send it as the HTTP response.
	err = app.writeJSON(w, http.StatusOK, car, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	car, err := app.models.Cars.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	var input struct {
		Model       *string `json:"model"`
		Year        *int64  `json:"year"`
		Price       *int64  `json:"price"`
		Marka       *string `json:"marka"`
		Color       *string `json:"color"`
		Type        *string `json:"type"`
		Image       *string `json:"image"`
		Description *string `json:"description"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Model != nil {
		car.Model = *input.Model
	}
	if input.Year != nil {
		car.Year = *input.Year
	}
	if input.Price != nil {
		car.Price = *input.Price
	}
	if input.Marka != nil {
		car.Marka = *input.Marka
	}
	if input.Color != nil {
		car.Color = *input.Color
	}
	if input.Type != nil {
		car.Type = *input.Type
	}
	if input.Image != nil {
		car.Image = *input.Image
	}
	if input.Description != nil {
		car.Description = *input.Description
	}

	v := validator.NewToActivate()
	if data.ValidateCar(v, car); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Cars.Update(car)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, car, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteCarHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Cars.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "car successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listCarsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Model       string
		Year        int64
		Marka       string
		PriceMax    int64
		PriceMin    int64
		Color       string
		Type        string
		Description string
		data.Filters
	}
	v := validator.NewToActivate()
	qs := r.URL.Query()

	input.Model = app.readString(qs, "model", "")
	input.Year = app.readInt(qs, "year", 2023, v)
	input.Marka = app.readString(qs, "marka", "")
	input.PriceMax = app.readInt(qs, "price_max", 150000000, v)
	input.PriceMin = app.readInt(qs, "price_min", 0, v)
	input.Model = app.readString(qs, "model", "")
	input.Type = app.readString(qs, "type", "")
	input.Description = app.readString(qs, "description", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = 20
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "model", "year", "price", "marka", "-id", "-model", "-year", "-price", "-marka"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	keys := data.Keys{
		PriceMax: input.PriceMax,
		PriceMin: input.PriceMin,
	}

	if data.ValidateKeys(v, keys); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	cars, err := app.models.Cars.GetAll(input.Model, input.Year, input.Marka, input.PriceMax, input.PriceMin,
		input.Color, input.Type, input.Description, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, cars, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
