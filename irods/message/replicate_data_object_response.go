package message

import (
	"fmt"

	"github.com/cyverse/go-irodsclient/irods/common"
	"github.com/cyverse/go-irodsclient/irods/types"
)

// IRODSMessageReplicateDataObjectResponse stores data object replication response
type IRODSMessageReplicateDataObjectResponse struct {
	// empty structure
	Result int
}

// CheckError returns error if server returned an error
func (msg *IRODSMessageReplicateDataObjectResponse) CheckError() error {
	if msg.Result < 0 {
		return types.NewIRODSError(common.ErrorCode(msg.Result))
	}
	return nil
}

// FromMessage returns struct from IRODSMessage
func (msg *IRODSMessageReplicateDataObjectResponse) FromMessage(msgIn *IRODSMessage) error {
	if msgIn.Body == nil {
		return fmt.Errorf("cannot create a struct from an empty body")
	}

	msg.Result = int(msgIn.Body.IntInfo)
	return nil
}