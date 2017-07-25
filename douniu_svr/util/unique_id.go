package util

import (
	"math/rand"
	"time"
)

var uidRand *rand.Rand
var curSeq int32

func init()  {
	uidRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	curSeq = rand.Int31()
}

func UniqueId() uint64 {
	now := time.Now().Unix()
	result := now << 32 | int64(curSeq)
	curSeq++
	return uint64(result)
}