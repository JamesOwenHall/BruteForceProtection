package hitcounter

type StringDirection struct {
	hits        *StringMap
	cleanUpTime float64
	name        string
	windowSize  float64
	incAmount   float64
}

func NewStringDirection(name string, windowSize, maxHits, cleanUpTime, maxTracked float64) *StringDirection {
	return &StringDirection{
		hits:        NewStringMap(int64(maxTracked)),
		cleanUpTime: cleanUpTime,
		name:        name,
		windowSize:  windowSize,
		incAmount:   windowSize / maxHits,
	}
}

func (s *StringDirection) Name() string {
	return s.name
}

func (s *StringDirection) Hit(clock int32, val interface{}) bool {
	// Make sure we have a string
	value, ok := val.(string)
	if !ok {
		return false
	}

	// We need to use the lock to access the hits map.
	status := s.hits.Lock(value)
	defer s.hits.Unlock(value)

	// Status is nil when we've exceeded the max number of tracked values.
	if status == nil {
		return false
	}

	// We're only dealing with floats from here on.
	fClock := float64(clock)

	if status.FrontTile < fClock {
		// No recent hits
		status.FrontTile = fClock + s.incAmount
		status.IsBlocked = false
		return true
	} else if status.FrontTile > fClock+s.windowSize {
		// We've crossed the threshold, start blocking
		status.IsBlocked = true
		return false
	} else {
		// We haven't crossed the threshold yet, let's increment
		status.FrontTile += s.incAmount

		// Now that we've incremented, we may have crossed the threshold
		if status.FrontTile > fClock+s.windowSize {
			// We crossed the threshold, start blocking
			status.IsBlocked = true
			return false
		} else {
			// We're not over the threshold even after incrementing.  But it's
			// possible that we crossed it earlier, so let's make sure we're
			// not already blocking.
			return !status.IsBlocked
		}
	}
}

func (s *StringDirection) CleanUp(clock int32) {
	for i, m := range s.hits.Mutexes {
		m.Lock()

		for k := range s.hits.Shards[i] {
			if s.hits.Shards[i][k].FrontTile < float64(clock) {
				delete(s.hits.Shards[i], k)
			}
		}

		m.Unlock()
	}
}

func (s *StringDirection) CleanUpTime() float64 {
	return s.cleanUpTime
}
