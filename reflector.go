package ezapi

import (
	"errors"
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
	_EZAPI_TAG_ALIAS    = "alias"
	_EZAPI_TAG_DESC     = "desc"
)

// to this struct represents the reflected struct
type reflectedReq struct {
	typ reflect.Type

	// JSON Body
	jsonBodyType      reflect.Type
	jsonBodyFieldName string

	// Path Params
	pathParamsType      reflect.Type
	pathParams          []reflectedKeyVal
	pathParamsFieldName string

	// Query Params
	queryParamsType      reflect.Type
	queryParams          []reflectedKeyVal
	queryParamsFieldName string

	// Context Values
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
	// Field
	typ       reflect.Type
	fieldName string

	// Modifiers
	alias       string
	aliasIsSet  bool
	optional    bool
	description string
}

// helper function to reflect the request struct
func ReflectReq[T any]() reflectedReq {
	var v T
	var errs []error
	t := reflect.TypeOf(v)

	// Create a reflectedReq struct
	reflected := reflectedReq{
		typ: t,
	}

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
				reflected.pathParams, errs = reflectParams(field.Type)
			case _EZAPI_TAG_QUERY_PARAMS:
				reflected.queryParamsType = field.Type
				reflected.queryParamsFieldName = field.Name
				reflected.queryParams, errs = reflectParams(field.Type)
			case _EZAPI_TAG_CONTEXT:
				reflected.contextValuesType = field.Type
				reflected.contextValuesName = field.Name
				reflected.contextValues, errs = reflectParams(field.Type)
			}
		}

		// If there are errors, panic
		if len(errs) > 0 {
			for i, err := range errs {
				errStr := err.Error()
				errStr = strings.ReplaceAll(errStr, "\n", "\n| ")
				fmt.Printf("Error %d: %s\n\n", i+1, errStr)
			}
			panic(fmt.Sprintf("error reflecting request struct '%s'", t.Name()))
		}
	}

	return reflected
}

// helper function to reflect the parameters of a struct
func reflectParams(t reflect.Type) ([]reflectedKeyVal, []error) {
	var errs []error
	if t.Kind() != reflect.Struct {
		return nil, []error{ErrInvalidParamsType}
	}
	params := []reflectedKeyVal{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(_EZAPI_TAG_NAME)

		// If the field has no tag, skip. It is not a parameter
		if tag == "" {
			continue
		}

		param := reflectedKeyVal{
			typ:         field.Type,
			fieldName:   field.Name,
			alias:       field.Name, // TODO: add warning if alias is not set
			aliasIsSet:  false,
			optional:    false,
			description: fmt.Sprintf("The %s parameter", field.Name),
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
				var name, value string
				nameValue := strings.Split(tagValue, "=")
				if len(nameValue) == 1 && !param.aliasIsSet {
					param.alias = nameValue[0]
					param.aliasIsSet = true
					continue
				}

				if len(nameValue) >= 1 {
					name = nameValue[0]
				}
				if len(nameValue) >= 2 {
					value = nameValue[1]
				}

				if name == "" || value == "" {
					errs = append(errs, errors.Join(ErrInvalidTag, fmt.Errorf("should be in the format key=value, got '%s'", tagValue)))
					continue
				}
				switch name {
				case _EZAPI_TAG_ALIAS:
					param.alias = value
					param.aliasIsSet = true
				case _EZAPI_TAG_DESC:
					param.description = value
				default:
					errs = append(errs, fmt.Errorf("unknown tag value '%s'", name))
					continue
				}
			}
		}

		params = append(params, param)
	}

	return params, errs
}

var (
	ErrInvalidParamsType = errors.New("invalid type for parameters")
	ErrInvalidTag        = errors.New("invalid tag")
)
