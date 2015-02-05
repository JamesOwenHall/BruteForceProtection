package hitcounter

import (
	"github.com/JamesOwenHall/BruteForceProtection/core/server"
)

type HitCounter struct {
	clock *Clock
	*server.Server
}

func NewHitCounter(directions []Direction) *HitCounter {
	result := new(HitCounter)
	result.clock = NewClock()
	result.Server = server.New()

	for _, dir := range directions {
		result.Routes[dir.Name()] = func(val interface{}) server.Response {
			resp := server.Response{
				Valid: dir.Hit(result.clock.GetTime(), val),
			}

			return resp
		}
	}

	return result
}
