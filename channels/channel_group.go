package channels

import (
	"fmt"
	"sync"
)

const (
	MaxChannels     = 16
	ChannelMinValue = 1000
	ChannelMaxValue = 2000
	ChannelMidValue = (ChannelMinValue + ChannelMaxValue) / 2 //1500
	DeadZone        = 5
)

type ChannelGroup struct {
	lock     sync.RWMutex
	channels []int
}

func NewChannelGroup() *ChannelGroup {
	return &ChannelGroup{
		channels: make([]int, MaxChannels),
	}
}

func (g *ChannelGroup) SetChannel(channel int, value int, inputType string, min int, max int) {
	g.lock.Lock()
	defer g.lock.Unlock()

	channelValue := ChannelMidValue
	switch inputType {
	case "axis_full":
		channelValue = mapToRangeWithDeadzone(value, min, max, ChannelMinValue, ChannelMaxValue, DeadZone)
	case "axis_top":
		channelValue = mapToRangeWithDeadzone(value, min, max, ChannelMidValue, ChannelMaxValue, DeadZone)
	case "axis_bottom":
		channelValue = mapToRangeWithDeadzone(value, min, max, ChannelMinValue, ChannelMidValue, DeadZone)
		channelValue = ChannelMidValue - channelValue + ChannelMinValue //Flip bottom half
	case "hat":
		if value == max {
			channelValue = ChannelMaxValue
		} else if value == min {
			channelValue = ChannelMinValue
		} else {
			channelValue = ChannelMidValue
		}
	case "button":
		if value == max {
			channelValue = ChannelMaxValue
		} else {
			channelValue = ChannelMidValue
		}
	}

	g.channels[channel] = channelValue
}

func (g *ChannelGroup) GetChannel(channel int) (int, error) {
	if channel < 0 || channel >= MaxChannels {
		return 0, fmt.Errorf("error: channel value out of range (0-%d)", MaxChannels)
	}

	g.lock.RLock()
	defer g.lock.RUnlock()

	return g.channels[channel], nil
}

func mapToRangeWithDeadzone(value, min, max, minReturn, maxReturn, deadzone int) int {
	mappedValue := mapToRange(value, min, max, minReturn, maxReturn)
	if ChannelMidValue+deadzone > mappedValue && mappedValue > ChannelMidValue {
		return ChannelMidValue
	} else if ChannelMidValue-deadzone < mappedValue && mappedValue < ChannelMidValue {
		return ChannelMidValue
	} else {
		return mappedValue
	}
}

func mapToRange(value, min, max, minReturn, maxReturn int) int {
	mappedValue := (maxReturn-minReturn)*(value-min)/(max-min) + minReturn

	if mappedValue > maxReturn {
		return maxReturn
	} else if mappedValue < minReturn {
		return minReturn
	} else {
		return mappedValue
	}
}
