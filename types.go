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

// Unmarshal Error (400)
type DefaultUnmarshalError struct {
	Err error
}

func (e DefaultUnmarshalError) Error() string {
	return e.Err.Error()
}

func (e DefaultUnmarshalError) Render(ctx BaseContext) error {
	w := ctx.GetW()
	w.WriteHeader(400)
	_, err := w.Write([]byte("Error unmarshalling request: " + e.Error()))
	return err
}
