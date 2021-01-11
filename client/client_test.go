package client

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	baseURL = "http://some-url/"
	data    = `{"data":{"A":"a"}}`
)

var okHTTPResponse = http.Response{
	StatusCode: 200,
	Body:       ioutil.NopCloser(bytes.NewBufferString(`{"success":"Good"}`)),
	Header:     make(http.Header),
}

var noContentHTTPResponse = http.Response{
	StatusCode: 204,
	Body:       ioutil.NopCloser(bytes.NewBufferString(``)),
	Header:     make(http.Header),
}

var okNonJSONHTTPResponse = &http.Response{
	StatusCode: 200,
	Body:       ioutil.NopCloser(bytes.NewBufferString(`OK`)),
	Header:     make(http.Header),
}

var unmarshallableErrorHTTPResponse = &http.Response{
	StatusCode: 500,
	Body:       ioutil.NopCloser(bytes.NewBufferString("Some non json error body")),
	Header:     make(http.Header),
}

var form3ErrorHTTPResponse = http.Response{
	StatusCode: 400,
	Body:       ioutil.NopCloser(bytes.NewBufferString(`{"failure":"Bad"}`)),
	Header:     make(http.Header),
}

var successResponse *FakeSuccessResponse
var errorResponse *FakeFailureResponse

var req *http.Request

func TestMain(m *testing.M) {
	successResponse = &FakeSuccessResponse{}
	errorResponse = &FakeFailureResponse{}
	req = &http.Request{}
}

func TestClientSendAndConsumeOKResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &okHTTPResponse, nil
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}

	resp, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)

	assert.NoError(t, err)
	assert.Equal(t, okHTTPResponse, *resp)

	want := &FakeSuccessResponse{
		Success: "Good",
	}
	assert.Equal(t, want, successResponse)
}

func TestClientConsumeEmptyContentResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &noContentHTTPResponse, nil
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}

	resp, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)

	assert.NoError(t, err)
	assert.Equal(t, noContentHTTPResponse, *resp)

	want := &FakeSuccessResponse{}
	assert.Equal(t, want, successResponse)
}

func TestClientConsumeOKNonJSONResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return okNonJSONHTTPResponse, nil
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}
	_, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)

	assert.Errorf(t, err, "invalid character 'O' looking for beginning of value")
}

func TestClientConsumeAccountErrorResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return &form3ErrorHTTPResponse, nil
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}
	resp, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, form3ErrorHTTPResponse, *resp)

	if !reflect.DeepEqual(form3ErrorHTTPResponse, *resp) {
		t.Errorf("Unexpected http.Response. Wanted %v, got %v", form3ErrorHTTPResponse, *resp)
	}

	want := &FakeFailureResponse{
		Failure: "Bad",
	}
	assert.Equal(t, want, errorResponse)
}

func TestClientConsumeHTTPStatusErrorNonJsonResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return unmarshallableErrorHTTPResponse, nil
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}
	_, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)

	assert.Errorf(t, err, "invalid character 'S' looking for beginning of value")
}

func TestClientDoErrorResponse(t *testing.T) {
	mock := &HTTPClientMock{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("An expected error")
		},
	}
	client := Client{
		HTTPClient: mock,
		BaseURL:    baseURL,
	}
	_, err := client.SendAndConsume("/some-path", "HTTP_METHOD", data, successResponse, errorResponse)

	assert.Errorf(t, err, "An expected error")
}

type HTTPClientMock struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *HTTPClientMock) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	panic("You should provide a do func implementation")
}

type FakeSuccessResponse struct {
	Success string `json:"success"`
}

type FakeFailureResponse struct {
	Failure string `json:"failure"`
}
