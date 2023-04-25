package main

import (
	"bytes"
	"carMarket.dreamteam.kz/internal/data"
	"carMarket.dreamteam.kz/internal/jsonlog"
	"carMarket.dreamteam.kz/internal/mailer"
	"carMarket.dreamteam.kz/internal/validator"
	"context"
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

var getAppInstance application

func init() {
	getAppInstance = GetApp()
}

func GetApp() application {
	var cfg config
	cfg.port = 4000
	cfg.env = "development"
	cfg.db.dsn = "postgres://postgres:12345@localhost/carmarket?sslmode=disable"

	cfg.db.maxOpenConns = 25
	cfg.db.maxIdleConns = 25
	cfg.db.maxIdleTime = "15m"

	cfg.smtp.host = "sandbox.smtp.mailtrap.io"
	cfg.smtp.port = 2525
	cfg.smtp.username = "211428@astanait.edu.kz"
	cfg.smtp.password = "Rik_Aitu2024*"
	cfg.smtp.sender = "211428@astanait.edu.kz"

	logger := jsonlog.NewToActivate(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.NewToActivate(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	return *app
}

func TestValidateUserTableDriven(t *testing.T) {

	var tests = []struct {
		name     string
		input    *data.User
		key      string
		expected string
	}{
		{
			name: "correct_user",
			input: &data.User{
				Name:      "Rash Estebek",
				Email:     "rashestebek@gmail.com",
				Role:      "user",
				Activated: false,
			},
			key:      "",
			expected: "",
		},
		{
			name: "empty_name",
			input: &data.User{
				Name:      "",
				Email:     "empty_name@gmail.com",
				Role:      "user",
				Activated: false,
			},
			key:      "name",
			expected: "must be provided",
		},
		{
			name: "long_name",
			input: &data.User{
				Name:      "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA BBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
				Email:     "long_name@gmail.com",
				Role:      "user",
				Activated: false,
			},
			key:      "name",
			expected: "must not be more than 50 bytes long",
		},
		{
			name: "empty_role",
			input: &data.User{
				Name:      "Abdurakhym Sayrambay",
				Email:     "abdu@gmail.com",
				Role:      "",
				Activated: false,
			},
			key:      "role",
			expected: "must be provided",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			tst.input.Password.Set("12345678")
			val := validator.NewToActivate()
			data.ValidateUser(val, tst.input)
			if val.Errors[tst.key] != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestValidatePermittedValueTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "matches_symbol",
			value:    "a",
			expected: true,
		},
		{
			name:     "wrong_symbol",
			value:    "b",
			expected: false,
		},
	}

	permittedValues := []string{"a", "c", "r"}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			val := validator.NewToActivate()
			val.Check(validator.PermittedValue(tst.value, permittedValues...), "size", "invalid size")
			if val.Valid() != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestValidateCarTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    *data.Car
		key      string
		expected string
	}{
		{
			name: "correct_car",
			input: &data.Car{
				Model:       "E200",
				Year:        2007,
				Price:       300000,
				Marka:       "Mercedes-Benz",
				Color:       "black",
				Type:        "sedan",
				Image:       "https://www.google.com/imgres?imgurl=https%3A%2F%2Fc0.carsie",
				Description: "good car",
			},
			key:      "",
			expected: "",
		},
		{
			name: "empty_model",
			input: &data.Car{
				Model:       "",
				Year:        2020,
				Price:       450000,
				Marka:       "Toyota",
				Color:       "blue",
				Type:        "sedan",
				Image:       "https://www.google.com/imgres?imgurl=https%3A%2F%2Fc0.carsie",
				Description: "fast car",
			},
			key:      "model",
			expected: "must be provided",
		},
		{
			name: "low_year",
			input: &data.Car{
				Model:       "E200",
				Year:        1750,
				Price:       150000,
				Marka:       "BMW",
				Color:       "white",
				Type:        "sedan",
				Image:       "https://www.google.com/imgres?imgurl=https%3A%2F%2Fc0.carsie",
				Description: "good car",
			},
			key:      "year",
			expected: "must be greater than 1800",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			val := validator.NewToActivate()
			data.ValidateCar(val, tst.input)
			if val.Errors[tst.key] != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestValidateMarkaTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    *data.Marka
		key      string
		expected string
	}{
		{
			name: "correct_marka",
			input: &data.Marka{
				Name:     "Mercedes",
				Producer: "Germany",
				Logo:     "https://group-media.mercedes-benz.com/marsMediaSite/Thumbnail?oid=46637120&version=-2&thumbnailVersion=3",
			},
			key:      "",
			expected: "",
		},
		{
			name: "empty_producer",
			input: &data.Marka{
				Name:     "Audi",
				Producer: "",
				Logo:     "https://group-media.mercedes-benz.com/marsMediaSite/Thumbnail?oid=46637120&version=-2&thumbnailVersion=3",
			},
			key:      "producer",
			expected: "must be provided",
		},
		{
			name: "empty_logo",
			input: &data.Marka{
				Name:     "Audi",
				Producer: "France",
				Logo:     "",
			},
			key:      "logo",
			expected: "must be provided",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			val := validator.NewToActivate()
			data.ValidateMarka(val, tst.input)
			if val.Errors[tst.key] != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestReadJSONTableDriven(t *testing.T) {
	type inputStruct struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var tests = []struct {
		name     string
		input    inputStruct
		expected string
	}{
		{
			name: "correct_json",
			input: inputStruct{
				Name:     "Erlan Mukhtarov",
				Email:    "erlan@gmail.com",
				Password: "erlaaa123",
				Role:     "user",
			},
			expected: "",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			bodyJSON, _ := json.Marshal(tst.input)
			bodyReader := bytes.NewReader(bodyJSON)
			w := httptest.NewRecorder()
			r, _ := http.NewRequest(http.MethodPost, "http://localhost:4000/v1/users", bodyReader)
			err := getAppInstance.readJSON(w, r, &tst.input)
			if err != nil && err.Error() != tst.expected {
				t.Errorf("Got an unexpected error: %v", err)
			}
		})
	}
}

func TestValidateKeysTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    *data.Keys
		key      string
		expected string
	}{
		{
			name: "correct_keys",
			input: &data.Keys{
				PriceMin: 100000,
				PriceMax: 500000,
			},
			key:      "",
			expected: "",
		},
		{
			name: "incorrect_keys",
			input: &data.Keys{
				PriceMin: 500000,
				PriceMax: 100000,
			},
			key:      "price",
			expected: "price_max must be greater than price_min",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			val := validator.NewToActivate()
			data.ValidateKeys(val, *tst.input)
			if val.Errors[tst.key] != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestValidateEmailTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		key      string
		expected string
	}{
		{
			name:     "correct_email",
			input:    "rik@gmail",
			key:      "",
			expected: "",
		},
		{
			name:     "empty_email",
			input:    "",
			key:      "email",
			expected: "must be provided",
		},
		{
			name:     "incorrect_email",
			input:    "rik",
			key:      "email",
			expected: "must be a valid email address",
		},
	}
	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			val := validator.NewToActivate()
			data.ValidateEmail(val, tst.input)
			if val.Errors[tst.key] != tst.expected {
				t.Errorf("We received an unexpected error: %v", val.Errors)
			}
		})
	}
}

func TestReadStringTableDrive(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "correct_key_marka",
			input:    "toyota",
			expected: "toyota",
		},
		{
			name:     "empty_key_marka",
			input:    "",
			expected: "",
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			url1, err := url.Parse("http://localhost:4000?marka=" + tst.input)
			urlValues := url1.Query()
			key := getAppInstance.readString(urlValues, "marka", "")
			if key != tst.expected || err != nil {
				t.Errorf("Expected value is %v, but got %v", tst.expected, key)
			}
		})
	}
}

func TestReadIntTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "correct_key_year",
			input:    "2007",
			expected: 2007,
		},
		{
			name:     "empty_key_year",
			input:    "",
			expected: 2003,
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			url1, err := url.Parse("http://localhost:4000?year=" + tst.input)
			urlValues := url1.Query()
			v := validator.NewToActivate()
			key := getAppInstance.readInt(urlValues, "year", 2003, v)
			if key != tst.expected || err != nil {
				t.Errorf("Expected value is %v, but got %v", tst.expected, key)
			}
		})
	}
}

func TestReadIDTableDriven(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "correct_key_id",
			input:    "1",
			expected: 1,
		},
		{
			name:     "empty_key_id",
			input:    "0",
			expected: 0,
		},
	}

	for _, tst := range tests {
		t.Run(tst.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/v1/cars", nil)

			params := httprouter.Params{
				{Key: "id", Value: tst.input},
			}
			ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
			req = req.WithContext(ctx)
			id, _ := getAppInstance.readIDParam(req)

			if id != tst.expected {
				t.Errorf("Expected value is %v, but got %v", tst.expected, id)
			}
		})
	}
}
