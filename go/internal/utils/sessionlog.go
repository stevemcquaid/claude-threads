package utils

// SessionLike is the minimal interface for objects that have a session ID.
type SessionLike interface {
	GetSessionID() string
}

// CreateSessionLog returns a factory function that creates session-scoped loggers.
// If session is nil, returns the base logger unchanged.
func CreateSessionLog(base *Logger) func(session SessionLike) *Logger {
	return func(session SessionLike) *Logger {
		if session == nil {
			return base
		}
		return base.ForSession(session.GetSessionID())
	}
}
