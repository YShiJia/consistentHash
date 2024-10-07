/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 21:33:31
 */

package test

import (
	"context"
	"github.com/gomodule/redigo/redis"
	"testing"
)

func TestRedisZset(t *testing.T) {
	conn, err := redis.DialContext(context.Background(),
		"tcp", "127.0.0.1:6379")
	if err != nil {
		t.Error(err.Error())
	}
	defer conn.Close()
	// redis.Values如果是获取的集合，没有数据返回空
	// 如果是获取的string， 没有数据返回nil
	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", "k1", 4, "+inf", "WITHSCORES"))
	t.Log(values, err)

	values, err = redis.Values(conn.Do("Get", "k2"))
	t.Log(values, err)

}
