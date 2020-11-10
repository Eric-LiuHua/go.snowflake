package idwork

import (
	"fmt"
)

/**
* workerIdBits             机器标识    动态消减
* datacenterIdBits         数据中心    动态消减
* sequenceBits             毫秒内自增的 12位 消减到10 位
* ----------------------------------------------------------------
* 确保输出的长度是需要的
*
* 核心代码为其IdWorker这个类实现，其原理结构如下，我分别用一个0表示一位，用—分割开部分的作用：
* 1||0---0000000000 0000000000 0000000000 0000000000 0 --- 00000 ---00000 ---000000000000
*
* 在上面的字符串中，第一位为未使用（实际上也可作为long的符号位），接下来的41位为毫秒级时间，
* 41位时间截(毫秒级)，注意，41位时间截不是存储当前时间的时间截，而是存储时间截的差值（当前时间截 - 开始时间截)
* 得到的值），这里的的开始时间截，一般是我们的id生成器开始使用的时间，由我们程序来指定的（如下下面程序IdWorker类的startTime属性）。41位的时间截，可以使用69年，年T = (1L << 41) / (1000L * 60 * 60 * 24 * 365) =
* 然后5位datacenter标识位，5位机器ID（并不算标识符，实际是为线程标识），
* 然后12位该毫秒内的当前毫秒内的计数，加起来刚好64位，为一个Long型。
* 这样的好处是，整体上按照时间自增排序，并且整个分布式系统内不会产生ID碰撞（由datacenter和机器ID作区分），
* 并且效率较高，经测试，snowflake每秒能够产生26万ID左右，完全满足需要。
* <p>
* 64位ID (42(毫秒)+5(机器ID)+5(业务编码)+12(重复累加))
 */

type Snowflake struct {
	//最大可以接受的参数左移量
	maxShift int64
	/**
	 * twepoch              时间起始标记点，作为基准，一般取系统的最近时间（一旦确定不能变动）
	 */
	twepoch          int64
	workerIdBits     int64
	datacenterIdBits int64
	sequenceBits     int64
	/**
	 * 最大值
	 * maxWorkerId              机器ID最大值
	 * maxDatacenterId          数据中心ID最大值
	 * sequenceMask             毫秒内自增最大数。后续求余需要
	 *
	 */
	maxWorkerId int64

	maxDatacenterId int64

	sequenceMask int64

	/**
	 * 二进制位置偏移量
	 * workerIdShift                机器ID偏左移位
	 * datacenterIdShift            数据中心ID左移位
	 * timestampLeftShift           时间毫秒左移位
	 */
	workerIdShift      int64
	datacenterIdShift  int64
	timestampLeftShift int64

	lastTimestamp int64
	sequence      int64
	workerId      int64
	datacenterId  int64
}

//系统自带的init，不允许参数和返回值。
func InitSnowflake(w int64, dc int64, s int64, twepoch int64) *Snowflake {
	tmp := &Snowflake{
		workerIdBits:     w,
		datacenterIdBits: dc,
		sequenceBits:     s,
		lastTimestamp:    -1,
		sequence:         0,
		workerId:         1,
		datacenterId:     1,
		twepoch:          twepoch,
	}
	/**
	 * 最大值
	 * maxWorkerId              机器ID最大值
	 * maxDatacenterId          数据中心ID最大值
	 * sequenceMask             毫秒内自增最大数。后续求余需要
	 *
	 */
	tmp.maxWorkerId = -1 ^ (-1 << tmp.workerIdBits)
	tmp.maxDatacenterId = -1 ^ (-1 << tmp.datacenterIdBits)
	tmp.sequenceMask = -1 ^ (-1 << tmp.sequenceBits)

	/**
	 * 二进制位置偏移量
	 * workerIdShift                机器ID偏左移位
	 * datacenterIdShift            数据中心ID左移位
	 * timestampLeftShift           时间毫秒左移位
	 */
	tmp.workerIdShift = tmp.sequenceBits
	tmp.datacenterIdShift = tmp.sequenceBits + tmp.workerIdBits
	tmp.timestampLeftShift = tmp.sequenceBits + tmp.workerIdBits + tmp.datacenterIdBits

	return tmp
}

//通过函数初始化默认值
func InitSnowflakeA() *Snowflake {
	return InitSnowflake(5, 5, 12, 1420041600000)
}

func (snowflake Snowflake) NewId(isDubber bool) int64 {
	var timestamp int64 = getTimestamp()
	var res int64 = -1
	if timestamp < snowflake.lastTimestamp {
		fmt.Printf("Clock moved backwards.  Refusing to generate id for %d milliseconds", (snowflake.lastTimestamp - timestamp))
	} else {
		if timestamp == snowflake.lastTimestamp {
			// 当前毫秒内，则+1，超过就是重复之前的，求余
			snowflake.sequence = (snowflake.sequence - 1) & snowflake.sequenceMask
			if snowflake.sequence == 0 {
				timestamp = tilNextMillis(snowflake.lastTimestamp)
			}
		} else {
			snowflake.sequence = 0
		}

		//重置最后时间
		snowflake.lastTimestamp = timestamp
		//ID偏移组合生成最终的ID，并返回ID
		res = (timestamp-snowflake.twepoch)<<snowflake.timestampLeftShift | snowflake.datacenterId<<snowflake.datacenterIdShift | snowflake.workerId<<snowflake.workerIdShift | snowflake.sequence
	}
	if isDubber {
		fmt.Printf("snowflake1.NewId() :%d ,timestamp:%d ,sequence:%d, sequenceMask:%d\n", res, timestamp, snowflake.sequence, snowflake.sequenceMask)
	}
	return res
}
