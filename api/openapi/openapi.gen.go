// Package openapi provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package openapi

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/runtime"
)

// PostCalendarsCalendarIdParams defines parameters for PostCalendarsCalendarId.
type PostCalendarsCalendarIdParams struct {
	Name string `form:"name" json:"name"`
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// PostCalendarsCalendarId request
	PostCalendarsCalendarId(ctx context.Context, calendarId string, params *PostCalendarsCalendarIdParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostSyncCalendarId request
	PostSyncCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteWatchCalendarId request
	DeleteWatchCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostWatchCalendarId request
	PostWatchCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostCalendarsCalendarId(ctx context.Context, calendarId string, params *PostCalendarsCalendarIdParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostCalendarsCalendarIdRequest(c.Server, calendarId, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostSyncCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostSyncCalendarIdRequest(c.Server, calendarId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteWatchCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteWatchCalendarIdRequest(c.Server, calendarId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostWatchCalendarId(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostWatchCalendarIdRequest(c.Server, calendarId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostCalendarsCalendarIdRequest generates requests for PostCalendarsCalendarId
func NewPostCalendarsCalendarIdRequest(server string, calendarId string, params *PostCalendarsCalendarIdParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "calendarId", runtime.ParamLocationPath, calendarId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/calendars/%s/", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	if params != nil {
		queryValues := queryURL.Query()

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "name", runtime.ParamLocationQuery, params.Name); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

		queryURL.RawQuery = queryValues.Encode()
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostSyncCalendarIdRequest generates requests for PostSyncCalendarId
func NewPostSyncCalendarIdRequest(server string, calendarId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "calendarId", runtime.ParamLocationPath, calendarId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sync/%s/", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteWatchCalendarIdRequest generates requests for DeleteWatchCalendarId
func NewDeleteWatchCalendarIdRequest(server string, calendarId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "calendarId", runtime.ParamLocationPath, calendarId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/watch/%s/", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPostWatchCalendarIdRequest generates requests for PostWatchCalendarId
func NewPostWatchCalendarIdRequest(server string, calendarId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "calendarId", runtime.ParamLocationPath, calendarId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/watch/%s/", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// PostCalendarsCalendarIdWithResponse request
	PostCalendarsCalendarIdWithResponse(ctx context.Context, calendarId string, params *PostCalendarsCalendarIdParams, reqEditors ...RequestEditorFn) (*PostCalendarsCalendarIdResponse, error)

	// PostSyncCalendarIdWithResponse request
	PostSyncCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*PostSyncCalendarIdResponse, error)

	// DeleteWatchCalendarIdWithResponse request
	DeleteWatchCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*DeleteWatchCalendarIdResponse, error)

	// PostWatchCalendarIdWithResponse request
	PostWatchCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*PostWatchCalendarIdResponse, error)
}

type PostCalendarsCalendarIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON201      *struct {
		Status *string `json:"status,omitempty"`
	}
	JSON401 *struct {
		Message *string `json:"message,omitempty"`
		Status  *string `json:"status,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r PostCalendarsCalendarIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostCalendarsCalendarIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostSyncCalendarIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Status *string `json:"status,omitempty"`
	}
	JSON404 *struct {
		Message *string `json:"message,omitempty"`
		Status  *string `json:"status,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r PostSyncCalendarIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostSyncCalendarIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteWatchCalendarIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Status *string `json:"status,omitempty"`
	}
	JSON404 *struct {
		Message *string `json:"message,omitempty"`
		Status  *string `json:"status,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r DeleteWatchCalendarIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteWatchCalendarIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostWatchCalendarIdResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Status *string `json:"status,omitempty"`
	}
	JSON404 *struct {
		Message *string `json:"message,omitempty"`
		Status  *string `json:"status,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r PostWatchCalendarIdResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostWatchCalendarIdResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostCalendarsCalendarIdWithResponse request returning *PostCalendarsCalendarIdResponse
func (c *ClientWithResponses) PostCalendarsCalendarIdWithResponse(ctx context.Context, calendarId string, params *PostCalendarsCalendarIdParams, reqEditors ...RequestEditorFn) (*PostCalendarsCalendarIdResponse, error) {
	rsp, err := c.PostCalendarsCalendarId(ctx, calendarId, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostCalendarsCalendarIdResponse(rsp)
}

// PostSyncCalendarIdWithResponse request returning *PostSyncCalendarIdResponse
func (c *ClientWithResponses) PostSyncCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*PostSyncCalendarIdResponse, error) {
	rsp, err := c.PostSyncCalendarId(ctx, calendarId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostSyncCalendarIdResponse(rsp)
}

// DeleteWatchCalendarIdWithResponse request returning *DeleteWatchCalendarIdResponse
func (c *ClientWithResponses) DeleteWatchCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*DeleteWatchCalendarIdResponse, error) {
	rsp, err := c.DeleteWatchCalendarId(ctx, calendarId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteWatchCalendarIdResponse(rsp)
}

// PostWatchCalendarIdWithResponse request returning *PostWatchCalendarIdResponse
func (c *ClientWithResponses) PostWatchCalendarIdWithResponse(ctx context.Context, calendarId string, reqEditors ...RequestEditorFn) (*PostWatchCalendarIdResponse, error) {
	rsp, err := c.PostWatchCalendarId(ctx, calendarId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostWatchCalendarIdResponse(rsp)
}

// ParsePostCalendarsCalendarIdResponse parses an HTTP response from a PostCalendarsCalendarIdWithResponse call
func ParsePostCalendarsCalendarIdResponse(rsp *http.Response) (*PostCalendarsCalendarIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostCalendarsCalendarIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 201:
		var dest struct {
			Status *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON201 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest struct {
			Message *string `json:"message,omitempty"`
			Status  *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	}

	return response, nil
}

// ParsePostSyncCalendarIdResponse parses an HTTP response from a PostSyncCalendarIdWithResponse call
func ParsePostSyncCalendarIdResponse(rsp *http.Response) (*PostSyncCalendarIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostSyncCalendarIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Status *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest struct {
			Message *string `json:"message,omitempty"`
			Status  *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}

// ParseDeleteWatchCalendarIdResponse parses an HTTP response from a DeleteWatchCalendarIdWithResponse call
func ParseDeleteWatchCalendarIdResponse(rsp *http.Response) (*DeleteWatchCalendarIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteWatchCalendarIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Status *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest struct {
			Message *string `json:"message,omitempty"`
			Status  *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}

// ParsePostWatchCalendarIdResponse parses an HTTP response from a PostWatchCalendarIdWithResponse call
func ParsePostWatchCalendarIdResponse(rsp *http.Response) (*PostWatchCalendarIdResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostWatchCalendarIdResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Status *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 404:
		var dest struct {
			Message *string `json:"message,omitempty"`
			Status  *string `json:"status,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON404 = &dest

	}

	return response, nil
}

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Create a new calendar
	// (POST /calendars/{calendarId}/)
	PostCalendarsCalendarId(ctx echo.Context, calendarId string, params PostCalendarsCalendarIdParams) error
	// Sync calendar information with local DB
	// (POST /sync/{calendarId}/)
	PostSyncCalendarId(ctx echo.Context, calendarId string) error
	// Stop watching a calendar
	// (DELETE /watch/{calendarId}/)
	DeleteWatchCalendarId(ctx echo.Context, calendarId string) error
	// Start watching a calendar
	// (POST /watch/{calendarId}/)
	PostWatchCalendarId(ctx echo.Context, calendarId string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// PostCalendarsCalendarId converts echo context to params.
func (w *ServerInterfaceWrapper) PostCalendarsCalendarId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "calendarId" -------------
	var calendarId string

	err = runtime.BindStyledParameterWithOptions("simple", "calendarId", ctx.Param("calendarId"), &calendarId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter calendarId: %s", err))
	}

	// Parameter object where we will unmarshal all parameters from the context
	var params PostCalendarsCalendarIdParams
	// ------------- Required query parameter "name" -------------

	err = runtime.BindQueryParameter("form", true, true, "name", ctx.QueryParams(), &params.Name)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter name: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostCalendarsCalendarId(ctx, calendarId, params)
	return err
}

// PostSyncCalendarId converts echo context to params.
func (w *ServerInterfaceWrapper) PostSyncCalendarId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "calendarId" -------------
	var calendarId string

	err = runtime.BindStyledParameterWithOptions("simple", "calendarId", ctx.Param("calendarId"), &calendarId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter calendarId: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostSyncCalendarId(ctx, calendarId)
	return err
}

// DeleteWatchCalendarId converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteWatchCalendarId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "calendarId" -------------
	var calendarId string

	err = runtime.BindStyledParameterWithOptions("simple", "calendarId", ctx.Param("calendarId"), &calendarId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter calendarId: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.DeleteWatchCalendarId(ctx, calendarId)
	return err
}

// PostWatchCalendarId converts echo context to params.
func (w *ServerInterfaceWrapper) PostWatchCalendarId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "calendarId" -------------
	var calendarId string

	err = runtime.BindStyledParameterWithOptions("simple", "calendarId", ctx.Param("calendarId"), &calendarId, runtime.BindStyledParameterOptions{ParamLocation: runtime.ParamLocationPath, Explode: false, Required: true})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter calendarId: %s", err))
	}

	// Invoke the callback with all the unmarshaled arguments
	err = w.Handler.PostWatchCalendarId(ctx, calendarId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.POST(baseURL+"/calendars/:calendarId/", wrapper.PostCalendarsCalendarId)
	router.POST(baseURL+"/sync/:calendarId/", wrapper.PostSyncCalendarId)
	router.DELETE(baseURL+"/watch/:calendarId/", wrapper.DeleteWatchCalendarId)
	router.POST(baseURL+"/watch/:calendarId/", wrapper.PostWatchCalendarId)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/+yUwW7bMAyGX0Xg2WjSrSffthYYciuwww5DDpzMJCpkSRXppUbgdx+oxEnaBMOKDsWA",
	"5SRDIX/yp75wAy4sItQbECeeoIYvMS49mVv0FBrM5msfrPl0P4MKflJmFwPUML26vprCUEFMFDA5qOFj",
	"uaogoaxYBSd2J8GTzfg5a4aJ/pYii54xUUZxMcwaqOE+sox1+XafUkQztiSUGerv2jPUpRBUELDVtu1x",
	"eKbHzmVqoJbcUQVsV9RicdknjWbJLixhGKqd2GNHuT+oleM1OnMN5hQDUzH/YXqth41BKBSrmJJ3tpid",
	"PLBOcXOkl7KOQtw2mwWlK1/0hG0q78KdtcQM1amH8Sb+eCArMOhVQ2yzS7J9r/1r2kwo1Jid2qLzvtd3",
	"vHlTvy0x45KeN7yviT4TNr2hJ8dyxkB11i/lHPMb3b6orJHctS3mXqPKKAyaQGsz4qMFcamU7VVgrnkT",
	"7oN9Dcj6v3kfhk/Zm/5L7JUFcuBtS9vN36btMDsTophF7ELznqjN7o7qPgetDGDsz+i+zW1xatZOVsZH",
	"i97cfT5CTzN22K1R7OqUu4Y8CZ2Sd1fuv2nSBT6ooUzCsMSUzm69/4pDickUnlxYGjy388q4YD5Uv1ls",
	"F7hewoVZLnDpFP6QrmEYfgUAAP//wzkqBXgKAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
