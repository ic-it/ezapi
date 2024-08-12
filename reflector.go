package ezapi

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	// tag name
	_EZAPI_TAG_NAME = "ezapi"

	// tag values
	_EZAPI_TAG_JSON_BODY    = "jsonBody"
	_EZAPI_TAG_PATH_PARAMS  = "path"
	_EZAPI_TAG_QUERY_PARAMS = "query"
	_EZAPI_TAG_CONTEXT      = "context"

	// tag values for params
	_EZAPI_TAG_OPTIONAL = "optional"
	_EZAPI_TAG_REQUIRED = "required"
)

// to this struct represents the reflected struct
type reflectedReq struct {
	jsonBodyType      reflect.Type
	jsonBodyFieldName string

	pathParamsType      reflect.Type
	pathParams          []reflectedKeyVal
	pathParamsFieldName string

	queryParamsType      reflect.Type
	queryParams          []reflectedKeyVal
	queryParamsFieldName string

	contextValuesType reflect.Type
	contextValues     []reflectedKeyVal
	contextValuesName string
}

func (rq reflectedReq) hasJSONBody() bool {
	return rq.jsonBodyType != nil
}

func (rq reflectedReq) hasPathParams() bool {
	return rq.pathParamsType != nil
}

func (rq reflectedReq) hasQueryParams() bool {
	return rq.queryParamsType != nil
}

func (rq reflectedReq) hasContextValues() bool {
	return rq.contextValuesType != nil
}

// reflected key value pair
type reflectedKeyVal struct {
	typ       reflect.Type
	fieldName string

	// Modifiers
	alias    string
	optional bool
}

// ReflectReq is a helper function that reflects the request struct
func ReflectReq[T any]() reflectedReq {
	var v T
	var err error
	t := reflect.TypeOf(v)

	// Create a reflectedReq struct
	reflected := reflectedReq{}

	// Iterate over the fields of the struct
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(_EZAPI_TAG_NAME)

		// If the field has the
		if tag != "" {
			switch tag {
			case _EZAPI_TAG_JSON_BODY:
				reflected.jsonBodyType = field.Type
				reflected.jsonBodyFieldName = field.Name
			case _EZAPI_TAG_PATH_PARAMS:
				reflected.pathParamsType = field.Type
				reflected.pathParamsFieldName = field.Name
				reflected.pathParams, err = reflectParams(field.Type)
			case _EZAPI_TAG_QUERY_PARAMS:
				reflected.queryParamsType = field.Type
				reflected.queryParamsFieldName = field.Name
				reflected.queryParams, err = reflectParams(field.Type)
			case _EZAPI_TAG_CONTEXT:
				reflected.contextValuesType = field.Type
				reflected.contextValuesName = field.Name
				reflected.contextValues, err = reflectParams(field.Type)
			}
		}

		if err != nil {
			panic(err)
		}
	}

	return reflected
}

// reflectParams is a helper function that reflects the params struct
func reflectParams(t reflect.Type) ([]reflectedKeyVal, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("params must be a struct %v", t)
	}
	params := []reflectedKeyVal{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(_EZAPI_TAG_NAME)

		// If the field has the
		if tag == "" {
			continue
		}
		param := reflectedKeyVal{
			typ:       field.Type,
			fieldName: field.Name,
			alias:     field.Name,
			optional:  false,
		}

		// tag values
		tagValues := strings.Split(tag, ",")
		for _, tagValue := range tagValues {
			switch tagValue {
			case _EZAPI_TAG_OPTIONAL:
				param.optional = true
			case _EZAPI_TAG_REQUIRED:
				param.optional = false
			default:
				param.alias = tagValue
			}
		}

		params = append(params, param)
	}

	return params, nil
}
