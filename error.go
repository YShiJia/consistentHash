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

	ErrVirtualNodeNotExistsCode = 40002
	ErrVirtualNodeNotExistsMsg  = "virtual node not exists"
)

var ErrNodeAlreadyExists = newError(ErrNodeAlreadyExistsCode, errors.New(ErrNodeAlreadyExistsMsg))
var ErrInvalidVirtualNodeID = newError(ErrInvalidVirtualNodeIDCode, errors.New(ErrInvalidVirtualNodeIDMsg))
var ErrVirtualNodeNotExists = newError(ErrVirtualNodeNotExistsCode, errors.New(ErrVirtualNodeNotExistsMsg))

func newError(code int64, err error) error {
	return fmt.Errorf("[err] code: %d  err: %w", code, err)
}
