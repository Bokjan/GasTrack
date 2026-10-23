package decode

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type sample struct {
	Name  string  `json:"name" validate:"required,min=2"`
	Year  int     `json:"year" validate:"omitempty,min=1900,max=2100"`
	Price float64 `json:"price" validate:"omitempty,gt=0"`
	Unit  string  `json:"unit" validate:"omitempty,oneof=L gal"`
}

func decodeBody(body string) error {
	r := httptest.NewRequest("POST", "/", strings.NewReader(body))
	var dst sample
	return JSON(r, &dst)
}

func TestJSONValidationMessages(t *testing.T) {
	cases := map[string]string{
		`{}`:                         "name is required",
		`{"name":"a"}`:               "name must be at least 2 characters",
		`{"name":"ab","year":1800}`:  "year must be at least 1900",
		`{"name":"ab","year":2200}`:  "year must be at most 2100",
		`{"name":"ab","price":-1}`:   "price must be greater than 0",
		`{"name":"ab","unit":"kWh"}`: "unit must be one of: L gal",
		`{"name":"ab","unknown":1}`:  "invalid JSON",
		`{"name":`:                   "invalid JSON",
		``:                           "request body is empty",
	}
	for body, want := range cases {
		err := decodeBody(body)
		if assert.Error(t, err, body) {
			assert.Contains(t, err.Error(), want, body)
		}
	}
	assert.NoError(t, decodeBody(`{"name":"ab","year":2000,"price":1.5,"unit":"L"}`))
}
