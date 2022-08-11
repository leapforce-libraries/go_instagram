package facebook

// ErrorResponse stores general API error response
//
type ErrorResponse struct {
	Error struct {
		Message         string `json:"message"`
		Type            string `json:"type"`
		Code            int    `json:"code"`
		ErrorSubcode    int    `json:"error_subcode"`
		IsTransient     int    `json:"is_transient"`
		ErrorUserTitle  string `json:"error_user_title"`
		ErrorUserMsg    string `json:"error_user_msg"`
		FacebookTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}
