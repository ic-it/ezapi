package goreapi

import (
	"encoding"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strconv"

	"github.com/google/uuid"
)

type unmarshaler[T any] func(
	body io.Reader,
	pathParams map[string]string,
	queryParams map[string][]string,
) (T, error)

// build unmarshaler for given reflectedReq
func BuildUnmarshaler[T any](reflected reflectedReq) unmarshaler[T] {
	// TODO: support pointers
	if reflected.hasPathParams() &&
		reflected.pathParamsType.Kind() != reflect.Struct {
		panic("pathParams must be a struct")
	}
	if reflected.hasQueryParams() &&
		reflected.queryParamsType.Kind() != reflect.Struct {
		panic("queryParams must be a struct")
	}

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
		v := reflect.New(reflected.pathParamsType).Elem()
		for _, param := range reflected.pathParams {
			field := v.FieldByName(param.fieldName)
			if !field.IsValid() {
				return nil, errInvalidField
			}

			value, ok := pathParams[param.alias]
			if !ok {
				if !param.optional {
					return nil, errMissingParam
				}
				continue
			}
			if value == "" && !param.optional {
				return nil, errMissingParam
			}

			unmarshaled, err := unmarshalType(param.typ, value)
			if err != nil {
				return nil, err
			}

			field.Set(reflect.ValueOf(unmarshaled))
		}
		return v.Interface(), nil
	}

	// query params deserializer
	queryParamsUnmarshaler := func(queryParams map[string][]string) (any, error) {
		v := reflect.New(reflected.queryParamsType).Elem()
		for _, param := range reflected.queryParams {
			field := v.FieldByName(param.fieldName)
			if !field.IsValid() {
				return nil, errInvalidField
			}

			values, ok := queryParams[param.alias]
			if !ok {
				if !param.optional {
					return nil, errMissingParam
				}
				continue
			}
			// TODO: handle multiple values
			if len(values) == 0 && values[0] == "" && !param.optional {
				return nil, errMissingParam
			}

			// TODO: handle multiple values
			unmarshaled, err := unmarshalType(param.typ, values[0])
			if err != nil {
				return nil, err
			}

			field.Set(reflect.ValueOf(unmarshaled))
		}
		return v.Interface(), nil
	}

	// return the unmarshaler
	return func(
		body io.Reader,
		pathParams map[string]string,
		queryParams map[string][]string,
	) (T, error) {
		var req reflect.Value
		{
			var t T
			req = reflect.New(reflect.TypeOf(t)).Elem()
		}

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

		return req.Interface().(T), nil
	}
}

func unmarshalType(typ reflect.Type, s string) (any, error) {
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
		v, err := unmarshalType(typ.Elem(), s)
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
	errInvalidField = errors.New("invalid field")
	errMissingParam = errors.New("missing param")
)
