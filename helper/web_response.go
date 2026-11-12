package helper

type WebResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Data   any    `json:"data,omitempty"`
	Errors any    `json:"errors,omitempty"`
}

func SuccessResponse(data any) WebResponse {
	return WebResponse{
		Code:   200,
		Status: "OK",
		Data:   data,
	}
}

func CreatedResponse(data any) WebResponse {
	return WebResponse{
		Code:   201,
		Status: "CREATED",
		Data:   data,
	}
}
