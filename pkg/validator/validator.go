package validator

import playgroundvalidator "github.com/go-playground/validator/v10"

type Validator struct {
	validate *playgroundvalidator.Validate
}

func New() *Validator {
	return &Validator{
		validate: playgroundvalidator.New(playgroundvalidator.WithRequiredStructEnabled()),
	}
}

func (v *Validator) Validate(value any) error {
	return v.validate.Struct(value)
}
