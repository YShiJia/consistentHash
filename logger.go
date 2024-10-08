/**
 * @author ysj
 * @email 2239831438@qq.com
 * @date 2024-09-27 00:28:23
 */

package csHash

type Logger interface {
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
}
