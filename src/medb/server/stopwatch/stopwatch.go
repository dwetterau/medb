package stopwatch

import (
	"log"
	"time"
)

type Stopwatch struct {
	name  string
	start time.Time
}

func Start(name string) Stopwatch {
	return Stopwatch{
		name,
		time.Now(),
	}
}

func (s Stopwatch) Stop(logger *log.Logger) {
	log.Printf("%s finished in %v", s.name, time.Since(s.start))
}
