package api_error

type JSONAPIError struct {
	Code         string `json:"code"`
	Msg          string `json:"msg"`
	ErrorDetails string `json:"error_details"`
}
