package ezapi

import "fmt"

func (r reflectedReq) String() string {
	return fmt.Sprintf(`EzAPI Reflected Request:
	NAME: %v
	JSON Body: %v
	Path Params: %v
	Query Params: %v
	Context Values: %v`,
		r.typ.Name(),
		r.jsonBodyType,
		r.pathParams,
		r.queryParams,
		r.contextValues,
	)
}

func (p reflectedKeyVal) String() string {
	return fmt.Sprintf("ReflectedKeyVal{ "+
		"Type: %v; "+
		"Field Name: %v; "+
		"Alias: %v; "+
		"Alias Is Set: %v; "+
		"Optional: %v; "+
		"Description: %v; "+
		"}",
		p.typ,
		p.fieldName,
		p.alias,
		p.aliasIsSet,
		p.optional,
		p.description,
	)
}
