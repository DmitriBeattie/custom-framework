package consumers

type errCode uint8

const (
	OK errCode = iota
	ErrCodeNotFoundQueue
	ErrCodeBadMessage
	ErrCodeFailedToPublish
)

