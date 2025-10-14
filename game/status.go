package game

// Status represents the allowed snake statuses.
type Status string

const (
	StatusOnline  Status = "online"
	StatusOffline Status = "offline"
	StatusLoading Status = "loading"
)
