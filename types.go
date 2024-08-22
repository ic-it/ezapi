package ezapi

type Renderable interface {
	Render(BaseContext) error
}

type Validatable interface {
	Validate(BaseContext) RespError
}

type OnUnmarshalError interface {
	OnUnmarshalError(BaseContext, error) RespError
}
type RespError interface {
	error
	Renderable
}
