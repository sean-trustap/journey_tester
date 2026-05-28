// Copyright 2026 Trustap. All rights reserved.
// Use of this source code is governed by an MIT
// licence that can be found in the LICENCE file.

package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"strings"

	"4d63.com/optional"
	"github.com/gofrs/uuid"
	journey_tester_context "github.com/trustap/journey_tester/pkg/journey_tester/context"
)

type APICall struct {
	Request  optional.Optional[string]
	Response optional.Optional[string]
}

func CallAPI(client *http.Client, c *journey_tester_context.Context, step Step) (*APICall, error) {
	resp, call, err := makeRequest(client, c, step)
	if err != nil {
		return call, fmt.Errorf("error making request to api: %w", err)
	}

	parsedResp, err := step.ParseResponse(resp)
	if err != nil {
		return call, fmt.Errorf("error checking api response: %w", err)
	}

	err = step.Extract(c, parsedResp)
	if err != nil {
		return call, fmt.Errorf("error extracting data: %w", err)
	}
	return call, nil
}

func makeRequest(client *http.Client, c *journey_tester_context.Context, step Step) (*http.Response, *APICall, error) {
	var res *http.Response
	call := &APICall{}

	req, err := step.NewRequest(c)
	if err != nil {
		return nil, call, fmt.Errorf("couldn't create new request: %w", err)
	}

	reqDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, call, fmt.Errorf("failed to get request as string: %w", err)
	}
	call.Request = optional.Of(string(reqDump))

	res, err = client.Do(req)
	if res == nil || err != nil {
		return nil, call, fmt.Errorf("failed to send request: %w", err)
	}

	resDump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return nil, call, fmt.Errorf("failed to get response as string: %w", err)
	}
	call.Response = optional.Of(string(resDump))

	return res, call, nil
}

func (config StepConfig) NewRequest(c *journey_tester_context.Context) (*http.Request, error) {
	return newRequest(config.Req)
}

func newRequest(reqConfig Req) (*http.Request, error) {
	var req *http.Request
	var body io.Reader
	if reqConfig.Body != nil {
		b, err := json.Marshal(reqConfig.Body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling configured req body: %w", err)
		}
		body = bytes.NewBuffer(b)
	}
	if reqConfig.Form != nil {
		form := url.Values{}
		for k, v := range *reqConfig.Form {
			form.Set(k, v)
		}
		if len(form) != 0 {
			body = strings.NewReader(form.Encode())
		}
	}
	var contentType optional.Optional[string]
	if reqConfig.File != nil {
		file, err := os.Open(reqConfig.File.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %v, err: %w", reqConfig.File.FileName, err)
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to stat file: %v, err: %w", reqConfig.File.FileName, err)
		}
		var fw io.Writer
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err = w.CreateFormFile(reqConfig.File.FieldName, fi.Name())
		if err != nil {
			return nil, fmt.Errorf("failed to create form file for request: %w", err)
		}
		_, err = io.Copy(fw, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy form file data for request: %w", err)
		}
		err = w.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to close form writer: %w", err)
		}

		contentType = optional.Of(w.FormDataContentType())
		body = &b
	}

	req, err := http.NewRequest(reqConfig.Method, reqConfig.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if reqConfig.Headers != nil {
		for key, value := range *reqConfig.Headers {
			req.Header.Set(key, value)
		}
	}

	// TODO This may result in unexpected behaviour if the test attempts to
	// provide a custom `Request-ID` header value, as the provided
	// `Request-ID` value will be overridden. This should ideally be
	// documented explicitly.
	req.Header.Set("Request-ID", newRequestID())

	if reqConfig.URLQuery != nil {
		urlValues := url.Values{}
		for k, v := range *reqConfig.URLQuery {
			urlValues.Set(k, v)
		}
		if len(urlValues) != 0 {
			req.URL.RawQuery = urlValues.Encode()
		}
	}

	if reqConfig.BasicAuth != nil {
		req.SetBasicAuth(reqConfig.BasicAuth.Username, reqConfig.BasicAuth.Password)
	}

	if value, ok := contentType.Get(); ok {
		req.Header.Set("Content-Type", value)
	}

	return req, nil
}

func newRequestID() string {
	id, err := uuid.NewV4()
	if err != nil {
		return fmt.Sprintf("%d", rand.Uint64())
	}
	return id.String()
}

func (config StepConfig) ParseResponse(res *http.Response) (*ParsedResponse, error) {
	return parseResponse(config.Resp, res, true)
}

func parseResponse(respConfig Resp, res *http.Response, allowExtraProperties bool) (*ParsedResponse, error) {
	if respConfig.Status != 0 && res.StatusCode != respConfig.Status {
		return nil, fmt.Errorf("response status doesn't match, expected: %v, actual: %v", respConfig.Status, res.Status)
	}

	defer res.Body.Close()
	rawBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if respConfig.Schema != nil && res.Body != nil {
		err = json.Unmarshal(rawBody, &respConfig.Schema)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshall resp into expected schema: %w", err)
		}
	}

	parsedResp := &ParsedResponse{
		RawBody: rawBody,
	}

	if respConfig.Body != nil {
		reflectActVal := reflect.Indirect(reflect.ValueOf(respConfig.Body))
		switch reflectActVal.Kind() {
		case reflect.Map:
			var jsonBody map[string]any
			err = json.Unmarshal(rawBody, &jsonBody)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshall resp into generic map[string]interface: %w", err)
			}
			parsedResp.JSONBody = &jsonBody

			expectedBody := respConfig.Body.(*map[string]any)

			err = Validate(allowExtraProperties, *expectedBody, jsonBody)
			if err != nil {
				return nil, fmt.Errorf("failed to validate response with expected data: %w", err)
			}

		case reflect.Slice:
			var arrayJSONBody []any
			err = json.Unmarshal(rawBody, &arrayJSONBody)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshall resp into generic []interface: %w", err)
			}
			parsedResp.ArrayJSONBody = &arrayJSONBody

			m := respConfig.Body.(*[]any)
			err = slicesMatch(allowExtraProperties, *m, arrayJSONBody)
			if err != nil {
				return nil, fmt.Errorf("failed to validate response with expected data: %w", err)
			}
		default:
			return nil, fmt.Errorf("reflection 'kind' of expected data must be map or slice, got: %v", reflectActVal.Kind())
		}
	}

	for _, respFunc := range respConfig.RespFuncs {
		err = respFunc.Func(rawBody)
		if err != nil {
			return nil, fmt.Errorf("failed to validate response with expected data: %w", err)
		}
	}

	return parsedResp, nil
}

func (config StepConfig) Extract(c *journey_tester_context.Context, parsedResp *ParsedResponse) error {
	return extract(config.Extracts, c, parsedResp)
}

func extract(extractsConfig Extracts, c *journey_tester_context.Context, parsedResp *ParsedResponse) error {
	for key, path := range extractsConfig.ObjectPaths {
		var data any
		if parsedResp.JSONBody != nil {
			data = parsedResp.JSONBody
		} else {
			data = parsedResp.ArrayJSONBody
		}
		val, err := GetStringValueAtPath(path, data)
		if err != nil {
			msg := "failed to extract value for '%s' at the following path: %v"
			return fmt.Errorf(msg, key, path)
		}
		c.Extracts.Set(key, val)
	}
	for key, extract := range extractsConfig.Extractors {
		v, err := extract(parsedResp.RawBody)
		if err != nil {
			msg := "failed to extract value for '%s' using custom function: %v"
			return fmt.Errorf(msg, key, err)
		}
		c.Extracts.Set(key, v)
	}
	return nil
}

type Step interface {
	NewRequest(c *journey_tester_context.Context) (*http.Request, error)
	ParseResponse(res *http.Response) (*ParsedResponse, error)
	Extract(c *journey_tester_context.Context, parsedResponse *ParsedResponse) error
}

type ParsedResponse struct {
	JSONBody      *map[string]any
	ArrayJSONBody *[]any
	RawBody       []byte
}
