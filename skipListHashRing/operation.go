/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:21:44
 */

package skipHashRing

type virtualNode struct {
	score  int64
	nodeID []string
	next   []*virtualNode
}
