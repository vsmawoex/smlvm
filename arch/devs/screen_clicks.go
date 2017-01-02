package devs

import (
	"fmt"
)

// ScreenClicks manages clicks on a screen.
type ScreenClicks struct {
	sender Sender
}

// NewScreenClicks creates new clicks handler.
func NewScreenClicks(s Sender) *ScreenClicks {
	return &ScreenClicks{sender: s}
}

// Click sends a click.
func (c *ScreenClicks) Click(line, col uint8) error {
	if line > ScreenHeight {
		return fmt.Errorf("line too big: %d", line)
	}
	if col > ScreenWidth {
		return fmt.Errorf("col too big: %d", col)
	}

	c.sender.Send([]byte{byte(line), byte(col)})
	return nil
}
