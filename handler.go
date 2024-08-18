package ezapi

import (
	"encoding/json"
	"net/http"
)

func H[T any, U any](handler func(Context[T]) (U, RespError)) http.HandlerFunc {
	reflected := ReflectReq[T]()
	unmarshler := BuildUnmarshaler[T](reflected)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := ezapiContext[T]{
			r: r,
			w: w,
		}

		qParams := map[string][]string{}
		pParams := map[string]string{}
		ctxVals := map[string]interface{}{}

		for _, p := range reflected.queryParams {
			vals, ok := r.URL.Query()[p.alias]
			if !ok {
				if !p.optional {
					http.Error(w, "missing query param: "+p.alias, http.StatusBadRequest)
					return
				}
				continue
			}
			qParams[p.alias] = vals
		}

		for _, p := range reflected.pathParams {
			pParams[p.alias] = r.PathValue(p.alias)
		}

		for _, p := range reflected.contextValues {
			ctxVals[p.alias] = r.Context().Value(p.alias)
		}

		req, err := unmarshler(r.Body, pParams, qParams, ctxVals)
		if err != nil {
			if onUnmarshalError, ok := any(req).(OnUnmarshalError); ok {
				err := onUnmarshalError.OnUnmarshalError(ctx, err)
				if err != nil {
					err := err.Render(ctx)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Validate the request
		if validatable, ok := any(req).(Validatable); ok {
			err := validatable.Validate(ctx)
			if err != nil {
				err := err.Render(ctx)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		ctx.req = req
		resp, handleErr := handler(ctx)
		if handleErr != nil {
			err := handleErr.Render(ctx)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
