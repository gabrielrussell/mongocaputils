package mongocaputils

import (
	"time"

	"github.com/gabrielrussell/mongocaputils/mongoproto"
)

type OpWithTime struct {
	mongoproto.OpRaw
	Seen time.Time
}

type orderedOps []OpWithTime

func (o orderedOps) Len() int {
	return len(o)
}

func (o orderedOps) Less(i, j int) bool {
	return o[i].Seen.Before(o[j].Seen)
}

func (o orderedOps) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o *orderedOps) Pop() interface{} {
	i := len(*o) - 1
	op := (*o)[i]
	*o = (*o)[:i]
	return op
}

func (o *orderedOps) Push(op interface{}) {
	*o = append(*o, op.(OpWithTime))
}
