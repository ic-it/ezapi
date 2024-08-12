package goreapi

type GoreError interface {
	error

	Render(BaseContext) error
}

type Validatable interface {
	Validate(BaseContext) GoreError
}

type OnUnmarshalError interface {
	OnUnmarshalError(BaseContext, error) GoreError
}
