package ezapi

type RespError interface {
	error

	Render(BaseContext) error
}

type Validatable interface {
	Validate(BaseContext) RespError
}

type OnUnmarshalError interface {
	OnUnmarshalError(BaseContext, error) RespError
}
