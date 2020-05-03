package wire

import (
	"testing"
	"time"
)

func TestUpdateReq_Encoding(t *testing.T) {
	updateReq := &UpdateReq{
		Name:      "testname",
		Timestamp: time.Unix(1234567, 0),
	}

	testMessageEncoding(t, "update_req", updateReq, &UpdateReq{})
}
