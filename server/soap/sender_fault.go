package soap

// SenderFaultError is the error channel for client mistakes: a handler
// returning it gets a SOAP Sender fault (HTTP 400) instead of the default
// Receiver fault — the wire-correct answer for unknown subscriptions,
// invalid durations, exceeded limits, and other request-side problems.
type SenderFaultError struct {
	// Reason is the short fault text shown to clients.
	Reason string
	// Detail carries the longer diagnostic.
	Detail string
}

// Error implements error.
func (e *SenderFaultError) Error() string {
	return e.Reason + ": " + e.Detail
}
