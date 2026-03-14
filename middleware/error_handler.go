package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProblemDetail follows RFC 7807 (Problem Details for HTTP APIs)
type ProblemDetail struct {
	Type     string      `json:"type"`
	Title    string      `json:"title"`
	Status   int         `json:"status"`
	Detail   string      `json:"detail"`
	Instance string      `json:"instance"`
	TraceID  string      `json:"trace_id,omitempty"`
	Extra    interface{} `json:"extra,omitempty"`
}

// RespondWithError sends an RFC 7807 formatted error response
func RespondWithError(c *gin.Context, statusCode int, errorType, detail string) {
	requestID := ""
	if val, exists := c.Get("request_id"); exists {
		requestID = val.(string)
	}

	titles := map[string]string{
		"missing_credentials":   "Missing Credentials",
		"invalid_auth_format":   "Invalid Authorization Format",
		"invalid_api_key":       "Invalid API Key",
		"rate_limit_exceeded":   "Rate Limit Exceeded",
		"invalid_url":           "Invalid URL",
		"url_not_found":         "URL Not Found",
		"duplicate_alias":       "Duplicate Custom Alias",
		"private_ip_blocked":    "Private IP Addresses Not Allowed",
		"unsupported_protocol":  "Unsupported Protocol",
		"url_too_long":          "URL Too Long",
		"internal_server_error": "Internal Server Error",
	}

	title := titles[errorType]
	if title == "" {
		title = http.StatusText(statusCode)
	}

	problem := ProblemDetail{
		Type:     fmt.Sprintf("https://api.url-shortener.dev/errors/%s", errorType),
		Title:    title,
		Status:   statusCode,
		Detail:   detail,
		Instance: c.Request.RequestURI,
		TraceID:  requestID,
	}

	c.JSON(statusCode, problem)
}

// RespondWithErrorAndExtra adds additional data to error response
func RespondWithErrorAndExtra(c *gin.Context, statusCode int, errorType, detail string, extra interface{}) {
	requestID := ""
	if val, exists := c.Get("request_id"); exists {
		requestID = val.(string)
	}

	titles := map[string]string{
		"missing_credentials":   "Missing Credentials",
		"invalid_auth_format":   "Invalid Authorization Format",
		"invalid_api_key":       "Invalid API Key",
		"rate_limit_exceeded":   "Rate Limit Exceeded",
		"invalid_url":           "Invalid URL",
		"url_not_found":         "URL Not Found",
		"duplicate_alias":       "Duplicate Custom Alias",
		"private_ip_blocked":    "Private IP Addresses Not Allowed",
		"unsupported_protocol":  "Unsupported Protocol",
		"url_too_long":          "URL Too Long",
		"internal_server_error": "Internal Server Error",
	}

	title := titles[errorType]
	if title == "" {
		title = http.StatusText(statusCode)
	}

	problem := ProblemDetail{
		Type:     fmt.Sprintf("https://api.url-shortener.dev/errors/%s", errorType),
		Title:    title,
		Status:   statusCode,
		Detail:   detail,
		Instance: c.Request.RequestURI,
		TraceID:  requestID,
		Extra:    extra,
	}

	c.JSON(statusCode, problem)
}
