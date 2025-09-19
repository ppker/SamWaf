package wafqueue

import (
	"SamWaf/common/zlog"
	"SamWaf/global"
	"SamWaf/innerbean"
	"SamWaf/waftask"
	"strconv"
	"sync/atomic"
	"time"
)

/*
*
处理Log队列信息
*/
func ProcessLogDequeEngine() {
	for {
		select {
		case <-global.GWAF_QUEUE_SHUTDOWN_SIGNAL:
			zlog.Info("日志队列处理协程收到关闭信号，正在退出...")
			return
		default:
			global.GWAF_MEASURE_PROCESS_DEQUEENGINE.WriteData(time.Now().UnixNano() / 1e6)
			if global.GDATA_CURRENT_CHANGE {
				//如果正在切换库 跳过
				zlog.Debug("正在切换数据库等待中队列")

			} else {
				var webLogArray []*innerbean.WebLog
				batchCount := 0
				for !global.GQEQUE_LOG_DB.Empty() {
					atomic.AddUint64(&global.GWAF_RUNTIME_LOG_PROCESS, 1) // 原子增加计数器
					weblogbean, ok := global.GQEQUE_LOG_DB.Dequeue()
					if !ok {
						continue
					}
					if weblogbean != nil {
						// 进行类型断言将其转为具体的结构
						if logValue, ok := weblogbean.(*innerbean.WebLog); ok {
							webLogArray = append(webLogArray, logValue)
							batchCount++
							if batchCount > int(global.GDATA_BATCH_INSERT) {
								break
							}
						} else {
							//插入其他类型内容
							global.GWAF_LOCAL_LOG_DB.Create(weblogbean)
						}
					}
				}
				if len(webLogArray) > 0 {
					zlog.Info("日志队列处理协程处理日志数量:" + strconv.Itoa(len(webLogArray)))
					if global.GCONFIG_LOG_PERSIST_ENABLED == 1 {
						global.GWAF_LOCAL_LOG_DB.CreateInBatches(webLogArray, len(webLogArray))
					}
					// 日志流做统计
					waftask.CollectStatsFromLogs(webLogArray)
					global.GNOTIFY_KAKFA_SERVICE.ProcessBatchLogs(webLogArray)
				}
			}
			time.Sleep(100 * time.Millisecond)
			global.GWAF_MEASURE_PROCESS_DEQUEENGINE.WriteData(time.Now().UnixNano() / 1e6)
		}
	}
}
