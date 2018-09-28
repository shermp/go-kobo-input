package koboin

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

type eventType uint16
type eventCode uint16

const (
	evAbs eventType = 0x03
	evKey eventType = 0x01
	evSyn eventType = 0x00
)

const (
	synReport            eventCode = 0x00
	synDropped           eventCode = 0x03
	synMTreport          eventCode = 0x02
	btnTouch             eventCode = 0x14a
	absX                 eventCode = 0x00
	absY                 eventCode = 0x01
	absMTposX            eventCode = 0x35
	absMTposY            eventCode = 0x36
	absMTPressure        eventCode = 0x3a
	absMTtouchWidthMajor eventCode = 0x30
)

type InputEvent struct {
	TimeSec  uint32
	TimeUsec uint32
	EvType   eventType
	EvCode   eventCode
	EvValue  int32
}

type TouchDevice struct {
	devFile    *os.File
	viewWidth  int
	viewHeight int
}

func New(inputPath string, vWidth, vHeight int) *TouchDevice {
	t := &TouchDevice{}
	f, err := os.OpenFile(inputPath, os.O_RDONLY, os.ModeDevice)
	if err != nil {
		fmt.Println("Could not open device for reading")
		return nil
	}
	t.devFile = f
	t.viewWidth = vWidth
	t.viewHeight = vHeight
	return t
}

func (t *TouchDevice) Close() {
	t.devFile.Close()
}

func (t *TouchDevice) getEvPacket() ([]InputEvent, error) {
	err := error(nil)
	// Keep looping until we have a complete event packet, signaled by a
	// SYN_REPORT event code
	evPacket := make([]InputEvent, 0)
	badPacket := false
	for {
		in := InputEvent{}
		if err = binary.Read(t.devFile, binary.LittleEndian, &in); err != nil {
			fmt.Println("binary.Read failed:", err)
			return evPacket, err
		}
		// Note, we have to check both event type and code together
		// which is why a switch statement isn't used here.
		if in.EvType == evSyn && in.EvCode == synDropped {
			// we need to ignore all packets up to, and including the next
			// SYN_REPORT
			badPacket = true
			evPacket = nil
			continue
		}
		if badPacket && in.EvType == evSyn && in.EvCode == synReport {
			// We encountered a SYN_DROPPED prevously. Return with an error
			return evPacket, errors.New("bad event packet")
		}
		if !badPacket {
			evPacket = append(evPacket, in)
			if in.EvType == evSyn && in.EvCode == synReport {
				// We have a complete event packet
				return evPacket, err
			}
		}
	}
}

func (t *TouchDevice) GetInput() (rx, ry int, err error) {
	err = nil
	x, y := -1, -1
	touchPressed := false
	touchReleased := false
	getEvAttempts := 0
	decodeEvAttempts := 0
	for {
		evPacket, err := t.getEvPacket()
		if err != nil {
			// we try again, increasing the attempts counter
			getEvAttempts++
			continue
		}
		if getEvAttempts > 4 {
			err = errors.New("failed to get touch packet")
			return x, y, err
		}
		// If we've got this far, we can reset the attempts counter
		getEvAttempts = 0
		// Now to decode the packet
		for _, e := range evPacket {
			switch e.EvType {
			// Some, but not all Kobo's report a BTN_TOUCH event
			case evKey:
				if e.EvCode == btnTouch {
					if e.EvValue == 1 {
						touchPressed = true
					} else {
						touchReleased = true
					}
				}
			case evAbs:
				switch e.EvCode {
				case absX:
					x = int(e.EvValue)
				case absY:
					y = int(e.EvValue)
				case absMTposX:
					x = int(e.EvValue)
				case absMTposY:
					y = int(e.EvValue)
				// Some kobo's seem to prefer using pressure to detect touch pressed/released
				case absMTPressure:
					if e.EvValue > 0 {
						touchPressed = true
					} else {
						touchReleased = true
					}
				// And others use the ABS_MT_WIDTH_MAJOR (and ABS_MT_TOUCH_MAJOR too, but those
				// are also used and set to zero on other kobo versions) instead :(
				case absMTtouchWidthMajor:
					if e.EvValue > 0 {
						touchPressed = true
					} else {
						touchReleased = true
					}
				}
			}
		}
		// We've decoded one packet. Do we need to continue?
		if x >= 0 && y >= 0 && touchPressed && touchReleased {
			// No, we have all the information we need
			break
		}
		// To ensure we never get caught in an infinite loop
		if decodeEvAttempts < 5 {
			decodeEvAttempts++
		} else {
			x, y, err = -1, -1, errors.New("unable to decode complete touch cycle")
			return x, y, err
		}
	}

	/*
		Coordinate rotation needs to happen.
		For reference, from FBInk, a clockwise rotation is as follows:

		rx = y
		ry = width - x - 1

		But we need to rotate counter-clockwise...
	*/
	ry = x
	rx = t.viewWidth - y + 1
	return rx, ry, nil
}
