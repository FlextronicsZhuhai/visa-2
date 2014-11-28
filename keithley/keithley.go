// Keithley RF Switch driver

package keithley

import (
	"fmt"

	vi "github.com/jpoirier/visa"
)

type Keithley struct {
	vi.Visa
	instr vi.Object
}

// RF_Switch Class - works with Keithley S46 RF Switch.
// Eight SPDT unterminated coaxial relays (2-pole) and four multi-pole unterminated coaxial relays.
// Relay A, multipole = Chan  1...6
// Relay B, multipole = Chan  7..12
// Relay C, multipole = Chan 13..18
// Relay D, multipole = Chan 19..24
// Relay 1..8, 2-pole = Chan 25..32
// Caution: Do not close more than one RF path per multiport switch.

// OpenGpib Opens a session to the specified resource.
func (k *Keithley) OpenGpib(rm vi.Session, ctrl, addr, mode, timeout uint32) (status vi.Status) {
	name := fmt.Sprintf("GPIB%d::%d", ctrl, addr)
	k.instr, status = rm.Open(name, mode, timeout)
	return
}

// OpenTcp Opens a session to the specified resource.
func (k *Keithley) OpenTcp(rm vi.Session, ip string, mode, timeout uint32) (status vi.Status) {
	name := fmt.Sprintf("TCPIP::%s::INSTR", ip)
	k.instr, status = rm.Open(name, mode, timeout)
	return
}

// Reset Resets the switch unit.
func (k *Keithley) Reset() (status vi.Status) {
	b := []byte("*RST")
	_, status = k.instr.Write(b, uint32(len(b)))
	return
}

// OpenChan Opens the specified channel. Where an open channel does not
// allow a signal to pass through.
func (k *Keithley) OpenChan(ch uint32) (status vi.Status) {
	b := fmt.Sprintf("OPEN (@%d)", ch)
	_, status = k.instr.Write([]byte(b), uint32(len(b)))
	return
}

// OpenAllChans Opens all channels.
func (k *Keithley) OpenAllChans() (status vi.Status) {
	b := []byte("OPEN:ALL")
	_, status = k.instr.Write(b, uint32(len(b)))
	return
}

// CloseChan Closes the specified channel. Note, All other channels on relay
// are opened first to prevent multiple closed relays leading to damage.
func (k *Keithley) CloseChan(ch uint32) (status vi.Status) {
	// Determine if ch is part of 2-port relay or multi-port relay (A..D)
	// Multi-Port Relay, A..D
	if ch > 0 && ch < 25 {
		// Open all ports on this relay
		for i := 1; i < 7; i++ {
			c := int((ch-1)/6)*6 + i
			k.OpenChan(uint32(c))
		}
	}
	b := fmt.Sprintf("CLOSE (@%d)", ch)
	_, status = k.instr.Write([]byte(b), uint32(len(b)))
	return
}

// ClosedChanList Returns a list of closed channels.
func (k *Keithley) ClosedChanList(ch uint32) (list string, status vi.Status) {
	// Returns list of closed channels.
	// RF Switch returns format '(@1,2,3)'.
	// If no channels closed, switch returns '(@)'.

	b := []byte("CLOSE?")
	_, status = k.instr.Write(b, uint32(len(b)))
	if status < vi.SUCCESS {
		return
	}
	buffer, _, status := k.instr.Read(100)
	if status < vi.SUCCESS {
		return
	} else {
		return string(buffer), status
	}
}

// if len(strClosedChans) > 3:  # Len always greater than 3 if channel in the list.
//     if strClosedChans.find(",") >= 0:
//         closedChans = strClosedChans[2:len(strClosedChans)-1].split(",")
//     else:
//         closedChans.append(strClosedChans[2:len(strClosedChans)-1])