/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:20:27
 */

package skipHashRing

import (
	"context"
	csHash "github.com/YShiJia/consistentHash"
)

type skipListHashRing struct {
	version int64
	head    *virtualNode
	// 每个节点对应的虚拟节点个数
	nodeNum        map[string]int64
	virtualNodeNum int64
	RealNodeNum    int64
}

func (s *skipListHashRing) Lock(ctx context.Context, expireSecond int64) error {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) Unlock(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) AddVirtualNode(ctx context.Context, score int64, nodeID string) (version int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) RemoveVirtualNode(ctx context.Context, score int64, nodeID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) GetVirtualNode(ctx context.Context, score int64) (nodeIDs *csHash.HashScore, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) FindDataToVirtualNode(ctx context.Context, dataScore int64) (virtualNodeID string, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) AddRealNode(ctx context.Context, nodeName string, replicas int64) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) GetRealNodes(ctx context.Context) (nodes map[string]int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) GetRealNode(ctx context.Context, nodeName string) (replicas int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) RemoveRealNode(ctx context.Context, nodeName string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) GetVersion(ctx context.Context) (version int64, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *skipListHashRing) SetVersion(ctx context.Context, version int64) (err error) {
	//TODO implement me
	panic("implement me")
}
