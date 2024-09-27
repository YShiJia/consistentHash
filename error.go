/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 12:36:44
 */

package csHash

import (
	"errors"
	"fmt"
)

const (
	ErrNodeAlreadyExistsCode = 40001
	ErrNodeAlreadyExistsMsg  = "node already exists"

	ErrInvalidVirtualNodeIDCode = 40002
	ErrInvalidVirtualNodeIDMsg  = "invalid virtual node id"
)

var ErrNodeAlreadyExists = errors.New(ErrNodeAlreadyExistsMsg)
var ErrInvalidVirtualNodeID = errors.New(ErrInvalidVirtualNodeIDMsg)

func NewError(code int64, err error) error {
	return fmt.Errorf("[err] code: %d msg :%w", code, err)
}
