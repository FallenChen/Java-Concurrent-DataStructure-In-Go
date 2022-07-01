package http2

import (
	"errors"
	"sync"
)

// section 5.2
type flowController struct {
	sync.RWMutex
	win,
	winLowerBound,
	winUpperBound,
	processedWin int
}

func (c *flowController) initialWindow() uint32 {
	c.RLock()
	win := c.winUpperBound
	c.RUnlock()

	return uint32(win)
}

func (c *flowController) window() int {
	c.RLock()
	win := c.win
	c.RUnlock()

	return win
}

func (c *flowController) updateWindow(delta int) error {
	if delta > 0 && maxInitialWindowSize-delta < c.win {
		return errors.New("window size overflow")
	}
	c.win += delta
	c.processedWin += delta
	c.winLowerBound = 0
	if delta < 0 {
		c.winLowerBound = delta
	}
	return nil
}
