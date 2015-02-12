package hitcounter

type Int32Direction struct {
	hits        *Int32Map
	cleanUpTime float64
	name        string
	windowSize  float64
	incAmount   float64
}

func NewInt32Direction(name string, windowSize, maxHits, cleanUpTime, maxTracked float64) *Int32Direction {
	return &Int32Direction{
		hits:        NewInt32Map(int64(maxTracked)),
		cleanUpTime: cleanUpTime,
		name:        name,
		windowSize:  windowSize,
		incAmount:   windowSize / maxHits,
	}
}

func (i *Int32Direction) Name() string {
	return i.name
}

func (i *Int32Direction) Hit(clock int32, val interface{}) bool {
	// JSON encodes numbers as float64, so we need to type assert, then cast.
	valueFloat, ok := val.(float64)
	if !ok {
		return false
	}
	value := int32(valueFloat)

	// We need to use the lock to access the hits map.
	status := i.hits.Lock(value)
	defer i.hits.Unlock(value)

	// Status is nil when we've exceeded the max number of tracked values.
	if status == nil {
		return false
	}

	// We're only dealing with floats from here on.
	fClock := float64(clock)

	if status.FrontTile < fClock {
		// No recent hits
		status.FrontTile = fClock + i.incAmount
		status.IsBlocked = false
		return true
	} else if status.FrontTile > fClock+i.windowSize {
		// We've crossed the threshold, start blocking
		status.IsBlocked = true
		return false
	} else {
		// We haven't crossed the threshold yet, let's increment
		status.FrontTile += i.incAmount

		// Now that we've incremented, we may have crossed the threshold
		if status.FrontTile > fClock+i.windowSize {
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

func (i *Int32Direction) CleanUp(clock int32) {
	for j, m := range i.hits.Mutexes {
		m.Lock()

		for k := range i.hits.Shards[j] {
			if i.hits.Shards[j][k].FrontTile < float64(clock) {
				delete(i.hits.Shards[j], k)
			}
		}

		m.Unlock()
	}
}

func (i *Int32Direction) CleanUpTime() float64 {
	return i.cleanUpTime
}
