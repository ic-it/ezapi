package ezapi

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/google/uuid"
)

type unmarshaler[T any] func(
	body io.Reader,
	pathParams map[string]string,
	queryParams map[string][]string,
	contextValues map[string]any,
) (T, error)

// build unmarshaler for given reflectedReq
func BuildUnmarshaler[T any](reflected reflectedReq) unmarshaler[T] {

	// json body unmarshaler
	jsonBodyUnmarshaler := func(body io.Reader) (any, error) {
		v := reflect.New(reflected.jsonBodyType).Interface()
		dec := json.NewDecoder(body)
		// dec.DisallowUnknownFields()
		err := dec.Decode(v)
		if err != nil {
			return nil, err
		}
		return v, nil
	}

	// path params deserializer
	pathParamsUnmarshaler := func(pathParams map[string]string) (any, error) {
		pathParamsType := reflected.pathParamsType
		isPtr := pathParamsType.Kind() == reflect.Ptr
		if pathParamsType.Kind() == reflect.Ptr {
			pathParamsType = pathParamsType.Elem()
		}
		v := reflect.New(pathParamsType).Elem()
		for _, param := range reflected.pathParams {
			debugErr := fmt.Errorf("alias: %s, field: %s", param.alias, param.fieldName)
			field := v.FieldByName(param.fieldName)
			if !field.IsValid() {
				return nil, errors.Join(ErrInvalidField, debugErr)
			}

			value, ok := pathParams[param.alias]
			if !ok {
				if !param.optional {
					return nil, errors.Join(ErrMissingPathParam, debugErr)
				}
				continue
			}
			if value == "" && !param.optional {
				return nil, errors.Join(ErrMissingPathParam, debugErr)
			}

			unmarshaled, err := unmarshalStrToType(param.typ, value)
			if err != nil {
				return nil, errors.Join(err, debugErr)
			}

			field.Set(reflect.ValueOf(unmarshaled))
		}
		if isPtr {
			return v.Addr().Interface(), nil
		}
		return v.Interface(), nil
	}

	// query params deserializer
	queryParamsUnmarshaler := func(queryParams map[string][]string) (any, error) {
		queryParamsType := reflected.queryParamsType
		isPtr := queryParamsType.Kind() == reflect.Ptr
		if queryParamsType.Kind() == reflect.Ptr {
			queryParamsType = queryParamsType.Elem()
		}
		v := reflect.New(queryParamsType).Elem()
		for _, param := range reflected.queryParams {
			debugErr := fmt.Errorf("alias: %s, field: %s", param.alias, param.fieldName)
			field := v.FieldByName(param.fieldName)
			if !field.IsValid() {
				return nil, errors.Join(ErrInvalidField, debugErr)
			}

			values, ok := queryParams[param.alias]
			if !ok {
				if !param.optional {
					return nil, errors.Join(ErrMissingQueryParam, debugErr)
				}
				continue
			}

			var unmarshaled any
			var err error
			if param.typ.Kind() == reflect.Slice {
				if len(values) == 0 {
					unmarshaled = reflect.Zero(param.typ).Interface()
				} else {
					unmarshaled, err = unmarshalSliceToType(param.typ, values)
					if err != nil {
						return nil, errors.Join(err, debugErr)
					}
				}
			} else if param.typ.Kind() == reflect.String {
				if len(values) == 0 {
					unmarshaled = ""
				} else {
					unmarshaled = values[0]
				}
			} else {
				return nil, errors.Join(ErrorUnsuppType, debugErr)
			}

			field.Set(reflect.ValueOf(unmarshaled))
		}
		if isPtr {
			return v.Addr().Interface(), nil
		}
		return v.Interface(), nil
	}

	contextValuesUnmarshaler := func(contextValues map[string]any) (any, error) {
		contextValuesType := reflected.contextValuesType
		isPtr := contextValuesType.Kind() == reflect.Ptr
		if contextValuesType.Kind() == reflect.Ptr {
			contextValuesType = contextValuesType.Elem()
		}
		v := reflect.New(contextValuesType).Elem()
		for _, param := range reflected.contextValues {
			debugErr := fmt.Errorf("alias: %s, field: %s", param.alias, param.fieldName)
			field := v.FieldByName(param.fieldName)
			if !field.IsValid() {
				return nil, errors.Join(ErrInvalidField, debugErr)
			}

			value, ok := contextValues[param.alias]
			if !ok {
				if !param.optional {
					return nil, errors.Join(ErrMissingContextValue, debugErr)
				}
				continue
			}
			if value == nil && !param.optional {
				return nil, errors.Join(ErrMissingContextValue, debugErr)
			}

			var unmarshaled any
			if str, ok := value.(string); ok {
				var err error
				unmarshaled, err = unmarshalStrToType(param.typ, str)
				if err != nil {
					return nil, errors.Join(err, debugErr)
				}
			} else {
				unmarshaled = value
			}

			// check type mismatch
			if reflect.TypeOf(unmarshaled) != param.typ {
				return nil, errors.Join(ErrTypeMismatch, debugErr,
					fmt.Errorf("expected type %v, got %v", param.typ, reflect.TypeOf(unmarshaled)),
				)
			}

			field.Set(reflect.ValueOf(unmarshaled))
		}
		// return v.Interface(), nil
		if isPtr {
			return v.Addr().Interface(), nil
		}
		return v.Interface(), nil
	}

	// return the unmarshaler
	return func(
		body io.Reader,
		pathParams map[string]string,
		queryParams map[string][]string,
		contextValues map[string]any,
	) (T, error) {
		req := reflect.New(reflected.typ).Elem()

		// unmarshal json body
		if reflected.hasJSONBody() {
			jsonBody, err := jsonBodyUnmarshaler(body)
			if err != nil {
				return req.Interface().(T), err
			}
			field := req.FieldByName(reflected.jsonBodyFieldName)
			if field.IsValid() {
				field.Set(reflect.ValueOf(jsonBody).Elem())
			}
		}

		// unmarshal path params
		if reflected.hasPathParams() {
			pathParamsStruct, err := pathParamsUnmarshaler(pathParams)
			if err != nil {
				return req.Interface().(T), err
			}
			field := req.FieldByName(reflected.pathParamsFieldName)
			if field.IsValid() {
				field.Set(reflect.ValueOf(pathParamsStruct))
			}
		}

		// unmarshal query params
		if reflected.hasQueryParams() {
			queryParamsStruct, err := queryParamsUnmarshaler(queryParams)
			if err != nil {
				return req.Interface().(T), err
			}
			field := req.FieldByName(reflected.queryParamsFieldName)
			if field.IsValid() {
				field.Set(reflect.ValueOf(queryParamsStruct))
			}
		}

		// unmarshal context values
		if reflected.hasContextValues() {
			contextValuesStruct, err := contextValuesUnmarshaler(contextValues)
			if err != nil {
				return req.Interface().(T), err
			}
			field := req.FieldByName(reflected.contextValuesName)
			if field.IsValid() {
				field.Set(reflect.ValueOf(contextValuesStruct))
			}
		}
		if reflected.isPtr {
			return req.Addr().Interface().(T), nil
		}
		return req.Interface().(T), nil
	}
}

func unmarshalSliceToType(typ reflect.Type, s []string) (any, error) {
	switch typ.Kind() {
	// SLICES
	case reflect.Slice:
		// Parse any other slice type
		newSlice := reflect.MakeSlice(typ, 0, 0)
		for _, part := range s {
			v, err := unmarshalStrToType(typ.Elem(), part)
			if err != nil {
				return nil, err
			}
			newSlice = reflect.Append(newSlice, reflect.ValueOf(v))
		}
		return newSlice.Interface(), nil
	default:
		panic("unsupported type")
	}
}

func unmarshalStrToType(typ reflect.Type, s string) (any, error) {
	switch typ.Kind() {
	// STRINGS
	case reflect.String:
		return s, nil
	// INTS
	case reflect.Int:
		return strconv.Atoi(s)
	case reflect.Int8:
		return strconv.ParseInt(s, 10, 8)
	case reflect.Int16:
		return strconv.ParseInt(s, 10, 16)
	case reflect.Int32:
		return strconv.ParseInt(s, 10, 32)
	case reflect.Int64:
		return strconv.ParseInt(s, 10, 64)
	// UINTS
	case reflect.Uint:
		return strconv.ParseUint(s, 10, 0)
	case reflect.Uint8:
		return strconv.ParseUint(s, 10, 8)
	case reflect.Uint16:
		return strconv.ParseUint(s, 10, 16)
	case reflect.Uint32:
		return strconv.ParseUint(s, 10, 32)
	case reflect.Uint64:
		return strconv.ParseUint(s, 10, 64)
	// FLOATS
	case reflect.Float64:
		return strconv.ParseFloat(s, 64)
	case reflect.Float32:
		return strconv.ParseFloat(s, 32)
	// BOOL
	case reflect.Bool:
		return strconv.ParseBool(s)
	// BYTES
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			return []byte(s), nil
		}
		panic("unsupported type")
	// STRUCT
	case reflect.Struct:
		// check if the type implements encoding.TextUnmarshaler
		if typ.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
			v := reflect.New(typ).Interface().(encoding.TextUnmarshaler)
			err := v.UnmarshalText([]byte(s))
			if err != nil {
				return nil, err
			}
			return v, nil
		}

		// check if the type implements json.Unmarshaler
		if typ.Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {
			v := reflect.New(typ).Interface().(json.Unmarshaler)
			err := v.UnmarshalJSON([]byte(s))
			if err != nil {
				return nil, err
			}
			return v, nil
		}

		panic("unsupported type")
	// POINTERS
	case reflect.Ptr:
		v, err := unmarshalStrToType(typ.Elem(), s)
		if err != nil {
			return nil, err
		}
		// create new pointer and set the value
		ptr := reflect.New(typ.Elem())
		ptr.Elem().Set(reflect.ValueOf(v))
		return ptr.Interface(), nil
	default:
		// parse UUID
		if typ == reflect.TypeOf(uuid.UUID{}) {
			return uuid.Parse(s)
		}

		panic("unsupported type")
	}
}

var (
	ErrInvalidField        = errors.New("invalid field")
	ErrMissingQueryParam   = errors.New("missing query param")
	ErrMissingPathParam    = errors.New("missing path param")
	ErrMissingContextValue = errors.New("missing context value")
	ErrTypeMismatch        = errors.New("type mismatch")
	ErrorUnsuppType        = errors.New("unsupported type")
)
