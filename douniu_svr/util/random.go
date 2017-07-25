package util

import (
	"math/rand"
	"time"
)

var r *rand.Rand

func init()  {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomN(n int) int {
	if n <= 0 {
		return 0
	}
	
	return r.Intn(n)
}

func Random(min, max int) int {
	if min >= max {
		return min
	}
	return min + r.Intn(max-min)
}

type Pool interface {
	Len() int
	Get(int) interface{}
	Remove(int)
}

func RandomTakeWay(pool Pool) (interface{}) {
	num := pool.Len()
	if num == 0 {
		return nil
	}
	takeWayIdx := Random(0, num)
	takeWay := pool.Get(takeWayIdx)
	pool.Remove(takeWayIdx)
	return takeWay
}
