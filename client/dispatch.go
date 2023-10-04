package client

import "github.com/soerenschneider/dyndns/internal/common"

type EventDispatch interface {
	Notify(msg *common.UpdateRecordRequest) error
}
