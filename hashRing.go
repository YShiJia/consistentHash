/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 18:53:03
 */

package csHash

import "context"

// 这里的HashRing接口，有一个版本号控制，方便使用者辨别前一个hash版本的数据，和一个现hash版本的数据。
// 每修改一次hash环，版本号加一
// 每一个score 可以存储多个nodeID，附带其版本号
type HashRing interface {
	// 锁住整个hash环
	Lock(ctx context.Context, expireSecond int64) error
	// 解锁hash环
	unlock(ctx context.Context) error
	// 在score标上添加节点
	AddNode(ctx context.Context, score int64, nodeID string) error
	// 删除score节点上的元素中数组的nodeID元素
	RemoveNode(ctx context.Context, score int64, nodeID string) error
	// 获取一个score节点上的元素中数组nodeID
	GetNode(ctx context.Context, score int64) (nodeIDs []string, _err error)
}
