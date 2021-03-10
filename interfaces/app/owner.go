package app

import (
	"github.com/satori/go.uuid"
)

type Owner interface {
	Ping(appID uuid.UUID) (isOwner bool, data interface{}, err error)
}
