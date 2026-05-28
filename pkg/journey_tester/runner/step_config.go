// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

// Deprecated: `StrictStepConfig` should be used instead, which performs strict
// response checking by default. `StrictStepConfig.AllowExtraProperties` can be
// used for temporary backwards-compatibility while migrating to
// `StrictStepConfig`, and can be used for requests to external resources, which
// don't require strict checking.
type StepConfig struct {
	Req      Req
	Resp     Resp
	Extracts Extracts
}

type Req struct {
	URL       string
	Method    string
	Headers   *map[string]string
	Body      any
	URLQuery  *map[string]string
	Form      *map[string]string
	BasicAuth *BasicAuth
	File      *File
}

type File struct {
	FieldName string
	FileName  string
}

type BasicAuth struct {
	Username string
	Password string
}

type Resp struct {
	Status int
	// `Schema` is a struct that contains the response data.
	Schema any
	// `Body` is the expected data, by field and value.
	Body      any
	RespFuncs []RespFunc
}

type RespFunc struct {
	Func func(body []byte) error
}

type Extracts struct {
	ObjectPaths map[string][]string
	Extractors  map[string]Extractor
}

type Extractor func(body []byte) (string, error)
