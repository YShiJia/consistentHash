/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-26 19:08:00
 */

package csHash

import "context"

// 迁移数据回调函数
type Migrator func(ctx context.Context, dataKeys map[string]struct{}, from, to string) error
