/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 18:53:03
 */

package csHash

import (
	"context"
)

/**
hashRing 有三种存储对象
1. HashRing， hashRing上面有2^31 - 1 个槽位，每个槽位有一个唯一的score，每个槽位上面可以有多个虚拟节点
2. RealNodeReplicas， 真实节点，以map的形式进行存储， key为RealNode名称， value为虚拟节点数量
3. ReadNodeData， 记录映射到真实节点中的数据，key为RealNode， value为数据列表
*/

type HashRing interface {
	// 锁住整个hash环
	Lock(ctx context.Context, expireSecond int64) error
	// 解锁hash环
	Unlock(ctx context.Context) error

	/* 向hash里面添加虚拟节点 */
	// 在score标上添加虚拟节点，如果节点已存在，直接返回
	AddVirtualNode(ctx context.Context, score int64, virtualNode string) (err error)
	// 删除score节点上的元素中数组的VirtualNode元素，不存在则直接返回
	RemoveVirtualNode(ctx context.Context, score int64, nodeID string) error
	// 获取一个score节点上的数据，若score不存在对应数据，返回零值
	GetVirtualNodes(ctx context.Context, score int64) (virtualNodes []string, err error)

	/*根据数据的score查询到对应槽位的虚拟节点，如果有多个虚拟节点，选择其一返回, 没有就返回空*/
	// 根据数据score，找到对应的节点，顺时针向下查找
	FindDataToVirtualNode(ctx context.Context, dataScore int64) (virtualNode string, err error)

	/* 对真实节点map进行操作 */
	// 添加真实节点列表，节点已存在，则更新
	AddRealNode(ctx context.Context, nodeName string, replicas int64) (err error)
	// 获取所有真实节点集合
	GetRealNodes(ctx context.Context) (nodes map[string]int64, err error)
	// 获取真实节点映射数量, nodeName不存在则报错
	GetRealNode(ctx context.Context, nodeName string) (replicas int64, err error)
	// 删除真实节点，不存在直接返回
	RemoveRealNode(ctx context.Context, nodeName string) (err error)

	/* 对数据进行操作 */
	// 添加数据到对应的集合中
	AddDataToRealNode(ctx context.Context, RealNode string, data map[string]struct{}) error
	// 获取真实节点对应的数据集合
	GetDataFromRealNode(ctx context.Context, RealNode string) (data map[string]struct{}, err error)
	// 在数据集合中删除数据，不存在则直接返回
	RemoveDataFromRealNode(ctx context.Context, RealNode string, data map[string]struct{}) error

	//一些工具方法
	Floor(ctx context.Context, score int64) (string, int64, error)
	Ceiling(ctx context.Context, score int64) (string, int64, error)
}
