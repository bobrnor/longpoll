package longpoll

import (
	"time"

	seqqueue "git.nulana.com/bobrnor/seqqueue.git"
)

type Entry struct {
	queue *seqqueue.Queue
	ts    time.Time
}
