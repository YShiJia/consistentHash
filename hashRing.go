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
	Unlock(ctx context.Context) error

	// 在score标上添加节点，节点已存在，则报错
	AddVirtualNode(ctx context.Context, score int64, nodeID string) (version int64, err error)
	// 删除score节点上的元素中数组的VirtualNode元素，不存在则直接返回
	// 如果删除VirtualNode并且score下面仍然还有元素，需要保证score的VirtualNodeList首元素可用
	RemoveVirtualNode(ctx context.Context, score int64, nodeID string) (version int64, err error)
	// 获取一个score节点上的数据，若score不存在对应数据，返回零值
	GetVirtualNode(ctx context.Context, score int64) (nodeIDs HashNode, err error)

	// 根据数据score，找到对应的节点，顺时针向下查找
	FindDataToVirtualNode(ctx context.Context, dataScore int64) (virtualNodeID string, err error)

	// 设置真实节点列表，节点已存在，则报错
	AddRealNode(ctx context.Context, nodeName string, replicas int64) (err error)
	// 获取真实节点列表
	GetRealNodes(ctx context.Context) (nodes map[string]int64, err error)
	// 获取真实节点映射数量, nodeName不存在则报错
	GetRealNode(ctx context.Context, nodeName string) (replicas int64, err error)
	// 删除真实节点，不存在直接返回
	RemoveRealNode(ctx context.Context, nodeName string) (err error)

	//查看当前hash环的版本号
	GetVersion(ctx context.Context) (version int64, err error)
	//修改当前hash环的版本号
	SetVersion(ctx context.Context, version int64) (err error)
}

type HashNode struct {
	VirtualNodeIDs []string `json:"virtual_node_ids"`
	Version        int64    `json:"version"`
}
