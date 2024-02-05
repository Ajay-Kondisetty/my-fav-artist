package utils

type APIResponse struct {
	Code  int         `json:"code"`
	Data  interface{} `json:"data"`
	Error string      `json:"error"`
}

type Data map[string]interface{}

type AppError struct {
	Error  error
	Status int
}

// PrepareResponse Prepares response format.
// It returns APIResponse.
func PrepareResponse(data interface{}, err error, code int) APIResponse {
	r := APIResponse{
		Code: code,
	}

	if err == nil {
		r.Data = data
	} else {
		r.Error = err.Error()
	}

	return r
}
