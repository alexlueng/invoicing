package util

import (
	"errors"
	"strconv"
	"sync"
	"time"
)
const (
	workerBits uint8 = 10 // 每台机器(节点)的ID位数 10位最大可以有2^10=1024个节点
	numberBits uint8 = 12 // 表示每个集群下的每个节点，1毫秒内可生成的id序号的二进制位数 即每毫秒可生成 2^12-1=4096个唯一ID
	// 这里求最大值使用了位运算，-1 的二进制表示为 1 的补码，感兴趣的同学可以自己算算试试 -1 ^ (-1 << nodeBits) 这里是不是等于 1023
	workerMax   int64 = -1 ^ (-1 << workerBits) // 节点ID的最大值，用于防止溢出
	numberMax   int64 = -1 ^ (-1 << numberBits) // 同上，用来表示生成id序号的最大值
	timeShift   uint8 = workerBits + numberBits // 时间戳向左的偏移量
	workerShift uint8 = numberBits              // 节点ID向左的偏移量
	// 41位字节作为时间戳数值的话 大约68年就会用完
	// 假如你2010年1月1日开始开发系统 如果不减去2010年1月1日的时间戳 那么白白浪费40年的时间戳啊！
	// 这个一旦定义且开始生成ID后千万不要改了 不然可能会生成相同的ID
	//epoch int64 = 1525705533000 // 这个是我在写epoch这个变量时的时间戳(毫秒)
	epoch int64 = 1595385525000 // 这个是我在写epoch这个变量时的时间戳(毫秒)
)

type Worker struct {
	mu        sync.Mutex
	timestamp int64
	workerID  int64
	number    int64
}

func NewWorker(workerID int64) (*Worker, error) {
	if workerID < 0 || workerID > workerMax {
		return nil, errors.New("Worker ID execess of quantity")
	}
	return &Worker{
		timestamp: 0,
		workerID: workerID,
		number: 0,
	}, nil
}

func (w *Worker) GetID() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UnixNano()
	if w.timestamp == now {
		w.number++
		if w.number > numberMax {
			for now <= w.timestamp {
				now = time.Now().UnixNano() / 1e6
			}
		}
	} else {
		w.number = 0
		w.timestamp = now
	}

	ID := int64((now-epoch)<<timeShift | (w.workerID << workerShift) | (w.number))
	if ID < 0 {
		return -ID
	}

	return ID

}

func (w *Worker) GetOrderSN(com_id, user_id int64) string {
	return strconv.FormatInt(w.GetID(), 10) + strconv.Itoa(int(com_id)) + strconv.Itoa(int(user_id))
}