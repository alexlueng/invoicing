package util

import (
	"testing"
)

func TestWorker_GetID(t *testing.T) {
	idWorker, _ := NewWorker(0)

	t.Log(idWorker.GetID())
	t.Log(idWorker.GetOrderSN(88, 7))
}