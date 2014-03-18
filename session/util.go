package session

import (
	"github.com/siddontang/polaris/util"
)

func GenerateID() string {
	return util.NewUUID().HexString()
}
