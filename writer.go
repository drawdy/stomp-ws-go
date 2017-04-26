//
// Copyright © 2011-2017 Guy M. Allard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package stompngo

import (
	"bufio"
	"net"
	// "bytes"
	"strconv"
	"time"
)

/*
	Logical network writer.  Read wiredata structures from the communication
	channel, and put the frame on the wire.
*/
func (c *Connection) writer() {
writerLoop:
	for {
		select {
		case d := <-c.output:
			c.log("WTR_WIREWRITE start")
			c.wireWrite(d)
			c.log("WTR_WIREWRITE COMPLETE", d.frame.Command, d.frame.Headers,
				HexData(d.frame.Body))
			if d.frame.Command == DISCONNECT {
				break writerLoop // we are done with this connection
			}
		case _ = <-c.ssdc:
			c.log("WTR_WIREWRITE shutdown S received")
			break writerLoop
		case _ = <-c.wtrsdc:
			c.log("WTR_WIREWRITE shutdown W received")
			break writerLoop
		}
	} // of for
	//
	c.log("WTR_SHUTDOWN", time.Now())
}

/*
	Connection logical write.
*/
func (c *Connection) wireWrite(d wiredata) {
	f := &d.frame
	// fmt.Printf("WWD01 f:[%v]\n", f)
	switch f.Command {
	case "\n": // HeartBeat frame
		if c.dld.wde && c.dld.dns {
			_ = c.netconn.SetWriteDeadline(time.Now().Add(c.dld.wdld))
		}
		_, e := c.wtr.WriteString(f.Command)
		if e != nil {
			if e.(net.Error).Timeout() {
				if c.dld.dns {
					c.dld.dlnotify(e, true)
				}
			}
			d.errchan <- e
			return
		}
	default: // Other frames
		if e := f.writeFrame(c.wtr, c); e != nil {
			d.errchan <- e
			return
		}
		if e := c.wtr.Flush(); e != nil {
			d.errchan <- e
			return
		}
	}
	if e := c.wtr.Flush(); e != nil {
		d.errchan <- e
		return
	}
	//
	if c.hbd != nil {
		c.hbd.sdl.Lock()
		c.hbd.ls = time.Now().UnixNano() // Latest good send
		c.hbd.sdl.Unlock()
	}
	c.mets.tfw++                // Frame written count
	c.mets.tbw += f.Size(false) // Bytes written count
	//
	d.errchan <- nil
	return
}

/*
	Physical frame write to the wire.
*/
func (f *Frame) writeFrame(w *bufio.Writer, c *Connection) error {

	var sctok bool
	// Content type.  Always add it if the client does not suppress and does not
	// supply it.
	_, sctok = f.Headers.Contains(HK_SUPPRESS_CT)
	if !sctok {
		if _, ctok := f.Headers.Contains(HK_CONTENT_TYPE); !ctok {
			f.Headers = append(f.Headers, HK_CONTENT_TYPE,
				DFLT_CONTENT_TYPE)
		}
	}

	var sclok bool
	// Content length - Always add it if client does not suppress it and
	// does not supply it.
	_, sclok = f.Headers.Contains(HK_SUPPRESS_CL)
	if !sclok {
		if _, clok := f.Headers.Contains(HK_CONTENT_LENGTH); !clok {
			f.Headers = append(f.Headers, HK_CONTENT_LENGTH, strconv.Itoa(len(f.Body)))
		}
	}

	if c.Protocol() > SPL_10 && f.Command != CONNECT {
		for i := 0; i < len(f.Headers); i += 2 {
			f.Headers[i] = encode(f.Headers[i])
			f.Headers[i+1] = encode(f.Headers[i+1])
		}
	}
	if c.dld.wde && c.dld.dns {
		_ = c.netconn.SetWriteDeadline(time.Now().Add(c.dld.wdld))
	}
	wba := f.Bytes(sclok)
	c.log("WIRE WRITE START")
	n, e := w.Write(wba)
	if e == nil {
		return e // Clean return
	}
	if n < len(wba) {
		c.log("SHORT WRITE", n, len(wba), e)
	}
	ne, ok := e.(net.Error)
	if !ok {
		return e // Some other error
	}
	if ne.Timeout() {
		//c.log("is a timeout")
		if c.dld.dns {
			c.log("invoking write deadline callback")
			c.dld.dlnotify(e, true)
		}
	}
	return e
}
