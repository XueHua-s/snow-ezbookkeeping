package core

import "context"

// assistantContext captures request metadata without retaining Gin or its writer.
type assistantContext struct {
	context.Context
	requestID string
	locale    string
	clientIP  string
}

func (c *assistantContext) ClientIP() string        { return c.clientIP }
func (c *assistantContext) GetContextId() string    { return c.requestID }
func (c *assistantContext) GetClientLocale() string { return c.locale }

// NewAssistantContext preserves cancellation and values from the HTTP request.
// Metadata must be captured by the handler before starting asynchronous work.
func NewAssistantContext(parent context.Context, requestID, locale, clientIP string) Context {
	if parent == nil {
		parent = context.Background()
	}
	return &assistantContext{Context: parent, requestID: requestID, locale: locale, clientIP: clientIP}
}
