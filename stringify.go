package ezapi

import "fmt"

func (r reflectedReq) String() string {
	return fmt.Sprintf(
		"ReflectedReq{jsonBodyType: %v, jsonBodyFieldName: %v, pathParams: %v, pathParamsFieldName: %v, queryParams: %v, queryParamsFieldName: %v}",
		r.jsonBodyType,
		r.jsonBodyFieldName,
		r.pathParams,
		r.pathParamsFieldName,
		r.queryParams,
		r.queryParamsFieldName,
	)
}

func (p reflectedKeyVal) String() string {
	return fmt.Sprintf("ReflectedParam{typ: %v, fieldName: %v, alias: %v, optional: %v}",
		p.typ, p.fieldName, p.alias, p.optional)
}
