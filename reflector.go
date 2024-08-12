package goreapi

import (
	"fmt"
	"reflect"
	"strings"
)

/*

type CreateUserReqBody struct {
	Name     string `json:"name"`
	Age      int    `json:"age,omitempty"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Example of code we want to reflect:
type CreateUserReq struct {
	JSONBody CreateUserReqBody  `gore:"jsonBody"`
	PathParams struct {
		ID int `gore:"id"`
	} `gore:"path"`
	QueryParams struct {
		ID int `gore:"id"`
		Status string `gore:"status,optional"`
	} `gore:"query"`
}

func (r CreateUserReq) Validate(goreCtx GoreContext) error {
	log.Println("Validating")
	return nil
}

func (r CreateUserReq) OnParseError(ctx GoreContext, err error) error {
	log.Println(err)
	return err
}

*/

const (
	// tag name
	_GORE_TAG_NAME = "gore"

	// tag values
	_GORE_TAG_JSON_BODY    = "jsonBody"
	_GORE_TAG_PATH_PARAMS  = "path"
	_GORE_TAG_QUERY_PARAMS = "query"

	// tag values for params
	_GORE_TAG_OPTIONAL = "optional"
	_GORE_TAG_REQUIRED = "required"
)

// to this struct represents the reflected struct
type reflectedReq struct {
	jsonBodyType      reflect.Type
	jsonBodyFieldName string

	pathParamsType      reflect.Type
	pathParams          []reflectedParam
	pathParamsFieldName string

	queryParamsType      reflect.Type
	queryParams          []reflectedParam
	queryParamsFieldName string

	// interface implementations
	isValidatable      bool
	isOnUnmarshalError bool
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

// reflected params
type reflectedParam struct {
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
		tag := field.Tag.Get(_GORE_TAG_NAME)

		// If the field has the gore tag
		if tag != "" {
			switch tag {
			case _GORE_TAG_JSON_BODY:
				reflected.jsonBodyType = field.Type
				reflected.jsonBodyFieldName = field.Name
			case _GORE_TAG_PATH_PARAMS:
				reflected.pathParamsType = field.Type
				reflected.pathParamsFieldName = field.Name
				reflected.pathParams, err = reflectParams(field.Type)
			case _GORE_TAG_QUERY_PARAMS:
				reflected.queryParamsType = field.Type
				reflected.queryParamsFieldName = field.Name
				reflected.queryParams, err = reflectParams(field.Type)
			}
		}

		if err != nil {
			panic(err)
		}
	}

	// Check if the struct implements the Validate method
	if _, ok := any(v).(Validatable); ok {
		reflected.isValidatable = true
	}

	// Check if the struct implements the OnUnmarshalError method
	if _, ok := any(v).(OnUnmarshalError); ok {
		reflected.isOnUnmarshalError = true
	}

	return reflected
}

// reflectParams is a helper function that reflects the params struct
func reflectParams(t reflect.Type) ([]reflectedParam, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("params must be a struct %v", t)
	}
	params := []reflectedParam{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(_GORE_TAG_NAME)

		// If the field has the gore tag
		if tag == "" {
			continue
		}
		param := reflectedParam{
			typ:       field.Type,
			fieldName: field.Name,
			alias:     field.Name,
			optional:  false,
		}

		// tag values
		tagValues := strings.Split(tag, ",")
		for _, tagValue := range tagValues {
			switch tagValue {
			case _GORE_TAG_OPTIONAL:
				param.optional = true
			case _GORE_TAG_REQUIRED:
				param.optional = false
			default:
				param.alias = tagValue
			}
		}

		params = append(params, param)
	}

	return params, nil
}
