/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:08:00
 */

package csHash

import (
	"context"
	"errors"
	"math"
)

// 迁移数据回调函数
type Migrator func(ctx context.Context, dataKeys map[string]struct{}, from, to string) error

/* migrateIn 和 migrateOut 都是在 virtualScore 对应的 virtualNode 存在的情况下执行的 */

// 添加节点的回调函数
// from to 都是真实节点
func (c *ConsistentHash) migrateIn(ctx context.Context, virtualScore int64, realNode string) (data map[string]struct{}, fromRealNode, toRealNode string, _err error) {
	/**
	1. 获取 curVirtualScore 的前后节点 preVirtualNode 和 nextVirtualNode
	2. 获取 preVirtualScore 到 curVirtualScore 之间的数据集合 a 就是返回值data
	3. 删除 nextVirtualNode 对应的数据集合中的 a 部分
	*/
	if c.migrator == nil {
		return
	}

	curVirtualScore := virtualScore
	curRealNode := realNode
	nodes, err := c.hashRing.GetVirtualNodes(ctx, curVirtualScore)
	if err != nil {
		_err = err
		return
	}
	// 当前score不由virtualScore负责
	if len(nodes) > 1 {
		_err = nil
		return
	}

	nextVirtualNode, nextVirtualScore, err := c.hashRing.Ceiling(ctx, virtualScore)
	if err != nil {
		_err = err
		return
	}
	if nextVirtualNode == "" || nextVirtualScore == curVirtualScore {
		// 当前hash环上面没有其他节点， 无需迁移
		_err = nil
		return
	}
	nextRealNode, _, _ := parseVirtualNodeID(nextVirtualNode)
	//当前节点和下一个节点是同一个真实节点，不用迁移
	if nextRealNode == curRealNode {
		_err = nil
		return
	}

	preVirtualNode, preVirtualScore, err := c.hashRing.Floor(ctx, virtualScore)
	if err != nil {
		_err = err
		return
	}
	// 这个逻辑前面出现过了，本来没必要，还是写一下吧
	if preVirtualNode == "" || preVirtualScore == curVirtualScore {
		// hash环无节点或者只有一个节点
		_err = nil
		return
	}
	if preVirtualScore == nextVirtualScore {
		// 当前hash环上面只有一个节点， 无需迁移
		_err = nil
		return
	}

	begin := []int64{}
	end := []int64{}
	if preVirtualScore < curVirtualScore {
		// pre-cur-next 或者 next-pre-cur
		begin = append(begin, preVirtualScore+1)
		end = append(end, curVirtualScore)
	} else if preVirtualScore > curVirtualScore {
		// cur-next-pre
		if preVirtualScore < math.MaxInt32 {
			begin = append(begin, preVirtualScore+1)
			end = append(end, math.MaxInt32)
		}
		begin = append(begin, 0)
		end = append(end, curVirtualScore)
	}

	// 获取需要迁移的数据
	dataSet, err := c.getDataSet(ctx, begin, end, nextRealNode)
	if err != nil {
		_err = err
		return
	}
	// 删除 nextRealNode 的部分数据集合
	_ = c.hashRing.RemoveDataFromRealNode(ctx, nextRealNode, dataSet)
	// 添加 curRealNode 的数据集合
	_ = c.hashRing.AddDataToRealNode(ctx, curRealNode, dataSet)
	// 从 nextRealNode 迁移数据到 curRealNode
	return dataSet, nextRealNode, curRealNode, nil
}

// 移除节点的回调函数, migrateIn 的逆向操作
func (c *ConsistentHash) migrateOut(ctx context.Context, virtualScore int64, realNode string) (data map[string]struct{}, fromRealNode, toRealNode string, _err error) {
	/**
	1. 获取 curVirtualScore 的前后节点 preVirtualNode 和 nextVirtualNode
	2. 获取 preVirtualScore 到 curVirtualScore 之间的数据集合 a 就是返回值data
	3. 在 nextVirtualNode 对应的数据集合加上集合a
	*/
	if c.migrator == nil {
		return
	}

	curVirtualScore := virtualScore
	curRealNode := realNode
	nodes, err := c.hashRing.GetVirtualNodes(ctx, curVirtualScore)
	if err != nil {
		_err = err
		return
	}
	// 当前score不由virtualScore负责
	if len(nodes) > 1 {
		_err = nil
		return
	}

	nextVirtualNode, nextVirtualScore, err := c.hashRing.Ceiling(ctx, virtualScore)
	if err != nil {
		_err = err
		return
	}
	if nextVirtualNode == "" || nextVirtualScore == curVirtualScore {
		// 当前hash环上面没有其他节点， 无需迁移
		_err = nil
		return
	}
	nextRealNode, _, _ := parseVirtualNodeID(nextVirtualNode)
	//当前节点和下一个节点是同一个真实节点，不用迁移
	if nextRealNode == curRealNode {
		_err = nil
		return
	}

	preVirtualNode, preVirtualScore, err := c.hashRing.Floor(ctx, virtualScore)
	if err != nil {
		_err = err
		return
	}
	// 这个逻辑前面出现过了，本来没必要，还是写一下吧
	if preVirtualNode == "" || preVirtualScore == curVirtualScore {
		// hash环无节点或者只有一个节点
		_err = nil
		return
	}
	if preVirtualScore == nextVirtualScore {
		// 当前hash环上面只有一个节点， 无需迁移
		_err = nil
		return
	}

	begin := []int64{}
	end := []int64{}
	if preVirtualScore < curVirtualScore {
		// pre-cur-next 或者 next-pre-cur
		begin = append(begin, preVirtualScore+1)
		end = append(end, curVirtualScore)
	} else if preVirtualScore > curVirtualScore {
		// cur-next-pre
		if preVirtualScore < math.MaxInt32 {
			begin = append(begin, preVirtualScore+1)
			end = append(end, math.MaxInt32)
		}
		begin = append(begin, 0)
		end = append(end, curVirtualScore)
	}

	// 获取需要迁移的数据
	dataSet, err := c.getDataSet(ctx, begin, end, nextRealNode)
	if err != nil {
		_err = err
		return
	}
	// 删除当前节点 的部分数据集合
	_ = c.hashRing.RemoveDataFromRealNode(ctx, curRealNode, dataSet)
	// 添加 nextVirtualNode 的数据集合
	_ = c.hashRing.AddDataToRealNode(ctx, nextRealNode, dataSet)
	// 从 curRealNode 迁移数据到 nextVirtualNode
	return dataSet, curRealNode, nextRealNode, nil
}

// range : [begin, end], 边界能取到
func (c *ConsistentHash) getDataSet(ctx context.Context, begin []int64, end []int64, realNode string) (data map[string]struct{}, err error) {
	if len(begin) != len(end) {
		return nil, errors.New("begin and end length not equal")
	}
	dataKeys, err := c.hashRing.GetDataFromRealNode(ctx, realNode)
	if err != nil {
		return nil, err
	}

	res := make(map[string]struct{})
	for key := range dataKeys {
		keyScore := int64(c.encryptor.Encrypt(key))
		flag := false
		for i := 0; i < len(begin); i++ {
			if keyScore >= begin[i] && keyScore <= end[i] {
				flag = true
				break
			}
		}
		if flag {
			res[key] = struct{}{}
		}
	}
	return res, nil
}

// 执行所有的数据迁移任务,需要顺序执行，不然会有数据错误，如果实际需求不考虑这一点，需要自行在 Migrator 中实现并发
func (c *ConsistentHash) batchExecuteMigrator(migrateTasks []func()) {
	defer func() {
		if err := recover(); err != nil {

		}
	}()

	for _, migrateTask := range migrateTasks {
		migrateTask()
	}
}
