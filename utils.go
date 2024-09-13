package main

import (
	"math/rand"
	"time"
)

// generates a random data fomated using f any date in [days] days before current time
// Recomended parameters: randTime("2006-01-02", 356*20)
func randTime(f string, days int64) string {
	var (
		ran int64 = 1000_000_000 * 60 * 60 * 24 * days
		now       = time.Now().UnixNano()
	)
	now -= ran
	now += rand.Int63n(ran)
	return time.Unix(0, now).Format(f)
}
