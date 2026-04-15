package validate

import "github.com/go-playground/validator/v10"

// validate is a validator instance to validate the request data
var validate = validator.New()

// ValidateStruct validates the request data
func ValidateStruct(req any) error {
	return validate.Struct(req)
}