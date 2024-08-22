package ezapi

import (
	"encoding/json"
	"net/http"
)

type handlerOpts struct {
	//
	contentType                      string
	defaultUnmarshalErrorConstructor func(error) RespError
}

func newHandlerOpts() *handlerOpts {
	return &handlerOpts{
		contentType: "application/json",
		defaultUnmarshalErrorConstructor: func(err error) RespError {
			return DefaultUnmarshalError{Err: err}
		},
	}
}

type HandlerOpt func(*handlerOpts)

func ContentType(contentType string) HandlerOpt {
	return func(o *handlerOpts) {
		o.contentType = contentType
	}
}

func DefaultUnmarshalErrorConstructor(constructor func(error) RespError) HandlerOpt {
	return func(o *handlerOpts) {
		o.defaultUnmarshalErrorConstructor = constructor
	}
}

func H[T any, U any](handler func(Context[T]) (U, RespError), opts ...HandlerOpt) http.HandlerFunc {
	options := newHandlerOpts()

	for _, opt := range opts {
		opt(options)
	}

	reflected := ReflectReq[T]()
	unmarshler := BuildUnmarshaler[T](reflected)

	return func(w http.ResponseWriter, r *http.Request) {
		var req T
		var err error

		ctx := ezapiContext[T]{
			r: r,
			w: w,
		}

		qParams := map[string][]string{}
		pParams := map[string]string{}
		ctxVals := map[string]any{}

		for _, p := range reflected.queryParams {
			vals, ok := r.URL.Query()[p.alias]
			if !ok {
				if !p.optional {
					missParamErr := MissingPathParamError{Param: p.alias}
					if oue, ok := any(req).(OnUnmarshalError); ok {
						if err := oue.OnUnmarshalError(ctx, missParamErr); err != nil {
							if err := err.Render(ctx); err != nil {
								DefaultInternalError{Err: err}.Render(ctx)
							}
						}
						return
					} else {
						err := options.defaultUnmarshalErrorConstructor(missParamErr)
						if err := err.Render(ctx); err != nil {
							DefaultInternalError{Err: err}.Render(ctx)
						}
						return
					}
				}
				continue
			}
			qParams[p.alias] = vals
		}

		for _, p := range reflected.pathParams {
			pp := r.PathValue(p.alias)
			if pp == "" {
				if !p.optional {
					missParamErr := MissingPathParamError{Param: p.alias}
					if oue, ok := any(req).(OnUnmarshalError); ok {
						if err := oue.OnUnmarshalError(ctx, missParamErr); err != nil {
							if err := err.Render(ctx); err != nil {
								DefaultInternalError{Err: err}.Render(ctx)
							}
						}
						return
					} else {
						err := options.defaultUnmarshalErrorConstructor(missParamErr)
						if err := err.Render(ctx); err != nil {
							DefaultInternalError{Err: err}.Render(ctx)
						}
						return
					}
				}
				continue
			}
			pParams[p.alias] = pp
		}

		for _, p := range reflected.contextValues {
			ctxVals[p.alias] = r.Context().Value(p.alias)
		}

		req, err = unmarshler(r.Body, pParams, qParams, ctxVals)
		if err != nil {
			if oue, ok := any(req).(OnUnmarshalError); ok {
				if err := oue.OnUnmarshalError(ctx, err); err != nil {
					if err := err.Render(ctx); err != nil {
						DefaultInternalError{Err: err}.Render(ctx)
					}
				}
				return
			} else {
				err := options.defaultUnmarshalErrorConstructor(err)
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		// Validate the request
		// Validate query params
		if validatorCb := reflected.queryParamsValidatorCb; validatorCb != nil {
			if err := validatorCb(req, ctx); err != nil {
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		// Validate path params
		if validatorCb := reflected.pathParamsValidatorCb; validatorCb != nil {
			if err := validatorCb(req, ctx); err != nil {
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		// Validate context values
		if validatorCb := reflected.contextValidatorCb; validatorCb != nil {
			if err := validatorCb(req, ctx); err != nil {
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		// Validate the json body
		if validatorCb := reflected.jsonBodyValidatorCb; validatorCb != nil {
			if err := validatorCb(req, ctx); err != nil {
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		// Validate the request
		if validatable, ok := any(req).(Validatable); ok {
			if err := validatable.Validate(ctx); err != nil {
				if err := err.Render(ctx); err != nil {
					DefaultInternalError{Err: err}.Render(ctx)
				}
				return
			}
		}

		ctx.req = req
		resp, handleErr := handler(ctx)
		if handleErr != nil {
			err := handleErr.Render(ctx)
			if err != nil {
				DefaultInternalError{Err: err}.Render(ctx)
			}
			return
		}

		if textResp, ok := any(resp).(string); ok {
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(textResp))
		} else if renderable, ok := any(resp).(Renderable); ok {
			if err := renderable.Render(ctx); err != nil {
				DefaultInternalError{Err: err}.Render(ctx)
			}
		} else {
			w.Header().Set("Content-Type", options.contentType)
			if err := json.NewEncoder(w).Encode(resp); err != nil {
				DefaultInternalError{Err: err}.Render(ctx)
			}
		}
	}
}

type MissingQueryParamError struct {
	Param string
}

func (e MissingQueryParamError) Error() string {
	return "missing query param: " + e.Param
}

type MissingPathParamError struct {
	Param string
}

func (e MissingPathParamError) Error() string {
	return "missing path param: " + e.Param
}
