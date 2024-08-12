package goreapi

import (
	"encoding/json"
	"net/http"
)

func H[T any, U any](handler func(GoreContext[T]) (U, GoreError)) http.HandlerFunc {
	reflected := ReflectReq[T]()
	unmarshler := BuildUnmarshaler[T](reflected)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := goreContext[T]{
			r: r,
			w: w,
		}

		qParams := map[string][]string{}
		pParams := map[string]string{}

		for _, p := range reflected.queryParams {
			qParams[p.alias] = []string{r.URL.Query().Get(p.alias)}
		}

		for _, p := range reflected.pathParams {
			pParams[p.alias] = r.PathValue(p.alias)
		}

		req, err := unmarshler(r.Body, pParams, qParams)
		if err != nil {
			if reflected.isOnUnmarshalError {
				err := any(req).(OnUnmarshalError).OnUnmarshalError(ctx, err)
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

		if reflected.isValidatable {
			err := any(req).(Validatable).Validate(ctx)
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
