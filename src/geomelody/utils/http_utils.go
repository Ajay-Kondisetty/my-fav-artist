package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

var transport *http2.Transport

var httpClient *http.Client

type ExternalRequest struct {
	Name    string                 `json:"name"`
	URL     string                 `json:"url"`
	Type    string                 `json:"type"`
	Headers map[string]string      `json:"headers"`
	Params  map[string]string      `json:"params"`
	Body    map[string]interface{} `json:"body"`
	RawBody []byte                 `json:"raw_body"`

	ReqCtx context.Context
}

type ExternalResponseAdditional struct {
	StatusCode int
	Headers    map[string][]string
}

type ExternalJSONResponse struct {
	Response    map[string]interface{}
	ArrResponse []interface{}
	StrResponse string
	ExternalResponseAdditional
}

// GetAPIResponse calls an API.
// It returns response of type interface and an error.
func GetAPIResponse(reqCtx context.Context, name, url, method string, body map[string]interface{}, params, headers map[string]string) (interface{}, error) {
	req := ExternalRequest{
		Name:    name,
		URL:     url,
		Type:    method,
		Params:  params,
		Headers: headers,
	}

	if body != nil {
		req.Body = body
	}

	var err error
	if reqCtx, err = GetExternalAPIResponse(req, reqCtx); err != nil {
		return nil, err
	}
	respStr, _ := reqCtx.Value("api." + req.Name).(string)
	var resp APIResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetExternalAPIResponse calls external api and adds the api response to the request context.
// It returns the updated context and error.
func GetExternalAPIResponse(req ExternalRequest, reqCtx context.Context) (context.Context, error) {
	req.ReqCtx = reqCtx

	if resp, err := req.Do(); err != nil {
		return reqCtx, err
	} else if re, err := ParseAsJSON(resp); err != nil {
		return reqCtx, err
	} else if re.StatusCode < 200 || re.StatusCode > 299 {
		return reqCtx, errors.New(fmt.Sprintf("%v", re.Response["errors"]))
	} else {
		apiResp := APIResponse{
			Code: re.StatusCode,
			Data: re.Response,
		}
		j, _ := json.Marshal(apiResp)
		reqCtx = context.WithValue(reqCtx, "api."+req.Name, string(j))
		return reqCtx, nil
	}
}

// Do method execute the ExternalRequest.
func (r *ExternalRequest) Do() (*http.Response, error) {
	r.GetMockHeadersFromContext()
	if _, ok := r.Headers["x-mock-api"]; ok {
		return r.DoMock()
	}

	var resp *http.Response
	body, err := r.getRequestBody()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(r.Type, r.URL, body)
	if err != nil {
		return nil, err
	}

	for k, v := range r.Headers {
		req.Header.Add(k, v)
	}

	q := req.URL.Query()
	for k, v := range r.Params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err = httpClient.Get(req.URL.String())
	if err != nil {
		return nil, err
	}

	return resp, err
}

// getRequestBody checks the content type and returns the request body string accordingly.
// It returns the request body string.
func (r *ExternalRequest) getRequestBody() (io.Reader, error) {
	var contentType string
	for k, v := range r.Headers {
		if strings.ToLower(k) == "content-type" {
			contentType = strings.ToLower(v)
			break
		}
	}

	switch contentType {
	case "application/json":
		j, _ := json.Marshal(r.Body)
		return strings.NewReader(string(j)), nil
	default:
		return nil, errors.New("received unsupported content type")
	}
}

// ParseAsJSON parses the response as json.
// It returns the response and error.
func ParseAsJSON(response *http.Response) (*ExternalJSONResponse, error) {
	defer func() {
		_ = response.Body.Close()
	}()

	jResp, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if jResp == nil || string(jResp) == "" {
		jResp = []byte(`{}`)
	}

	var dResp map[string]interface{}
	var dArrResp []interface{}
	var dStrResp string
	if err := json.Unmarshal(jResp, &dResp); err != nil {
		if err := json.Unmarshal(jResp, &dArrResp); err != nil {
			if response.StatusCode < 200 || response.StatusCode >= 300 {
				dStrResp = string(jResp)
			} else {
				return nil, err
			}
		}
	}

	extResp := &ExternalJSONResponse{
		Response:    dResp,
		ArrResponse: dArrResp,
		StrResponse: dStrResp,
		ExternalResponseAdditional: ExternalResponseAdditional{
			StatusCode: response.StatusCode,
			Headers:    response.Header,
		},
	}

	return extResp, nil
}

// updateDurationFromEnv fetches given var from env and sets it to passed duration variable.
func updateDurationFromEnv(enVar string, rVar *time.Duration) {
	if val := os.Getenv(enVar); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			*rVar = d
		} else {
			log.Printf("Error parsing duration %v: %v", enVar, err)
		}
	}
}

func init() {
	//var idleConnTimeout = 30 * time.Second
	//var respHeaderTimeout = 5 * time.Second
	//var maxIdleConnections = 100
	//
	//updateDurationFromEnv("HTTP_IDLE_CONN_TIMEOUT", &idleConnTimeout)
	//updateDurationFromEnv("HTTP_RESPONSE_HEADER_TIMEOUT", &respHeaderTimeout)
	//
	//if val := os.Getenv("HTTP_MAX_IDLE_CONNS"); val != "" {
	//	if i, err := strconv.Atoi(val); err == nil {
	//		maxIdleConnections = i
	//	} else {
	//		log.Printf("Error parsing integer HTTP_MAX_IDLE_CONNS: %v", err)
	//	}
	//}

	transport = &http2.Transport{}

	httpClient = &http.Client{Transport: transport}
}
