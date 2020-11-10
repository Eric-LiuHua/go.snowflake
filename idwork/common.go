package idwork

import "time"

//下以毫秒
func tilNextMillis(last int64) int64 {
    var timestamp int64 = getTimestamp()
    for {
        if timestamp > last {
            break
        } else {
            timestamp = getTimestamp()
        }
    }
    return timestamp
}

//毫秒  time.Now().UnixNano() / 1e6
func getTimestamp() int64 {
    return time.Now().UnixNano() / 1e6
}
