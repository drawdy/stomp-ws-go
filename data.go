//
// Copyright © 2011-2013 Guy M. Allard
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed, an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package stompngo

import (
	"bufio"
	"log"
	"net"
	"sync"
	"time"
)

const (

	// Client generated commands.
	CONNECT     = "CONNECT"
	STOMP       = "STOMP"
	DISCONNECT  = "DISCONNECT"
	SEND        = "SEND"
	SUBSCRIBE   = "SUBSCRIBE"
	UNSUBSCRIBE = "UNSUBSCRIBE"
	ACK         = "ACK"
	NACK        = "NACK"
	BEGIN       = "BEGIN"
	COMMIT      = "COMMIT"
	ABORT       = "ABORT"

	// Server generated commands.
	CONNECTED = "CONNECTED"
	MESSAGE   = "MESSAGE"
	RECEIPT   = "RECEIPT"
	ERROR     = "ERROR"

	// Supported STOMP protocol definitions.
	SPL_10 = "1.0"
	SPL_11 = "1.1"
	SPL_12 = "1.2"
)

/*
	What this package currently supports.
*/
var supported = []string{SPL_10, SPL_11, SPL_12}

/*
	Headers definition, a slice of string.

	STOMP headers are key and value pairs.  See the specification for more
	information about STOMP frame headers.

	Key values are found at even numbered indices.  Values
	are found at odd numbered incices.  Headers are validated for an even
	number of slice elements.
*/
type Headers []string

/*
	Message is a STOMP Message, consisting of: a STOMP command; a set of STOMP
	Headers; and a message body(payload), which is possibly empty.
*/
type Message struct {
	Command string
	Headers Headers
	Body    []uint8
}

/*
	Frame is an alternate name for a Message.
*/
type Frame Message

/*
	MessageData passed to the client, containing: the Message; and an Error
	value which is possibly nil.

	Note that this has no relevance on whether a MessageData.Message.Command
	value contains an "ERROR" generated by the broker.
*/
type MessageData struct {
	Message Message
	Error   error
}

/*
	This is outbound on the wire.
*/
type wiredata struct {
	frame   Frame
	errchan chan error
}

/*
	Connection is a representation of a STOMP connection.
*/
type Connection struct {
	ConnectResponse   *Message           // Broker response (CONNECTED/ERROR) if physical connection successful.
	DisconnectReceipt MessageData        // If receipt requested on DISCONNECT.
	MessageData       <-chan MessageData // Inbound data for the client.
	connected         bool
	session           string
	protocol          string
	input             chan MessageData
	output            chan wiredata
	netconn           net.Conn
	subs              map[string]chan MessageData
	subsLock          sync.Mutex
	wsd               chan bool // writer shutdown
	rsd               chan bool // reader shutdown
	hbd               *heartbeat_data
	wtr               *bufio.Writer
	rdr               *bufio.Reader
	Hbrf              bool // Indicates a heart beat read/receive failure, which is possibly transient.  Valid for 1.1+ only.
	Hbsf              bool // Indicates a heart beat send failure, which is possibly transient.  Valid for 1.1+ only.
	logger            *log.Logger
	mets              *metrics // Client metrics
	scc               int      // Subscribe channel capacity
}

/*
	Error definition.
*/
type Error string

/*
	Error constants.
*/
const (
	// ERROR Frame returned by broker on connect.
	ECONERR = Error("broker returned ERROR frame, CONNECT")

	// ERRRORs for Headers.
	EHDRLEN  = Error("unmatched headers, bad length")
	EHDRUTF8 = Error("header string not UTF8")
	EHDRNIL  = Error("headers can not be nil")

	// ERRORs for response to CONNECT.
	EUNKFRM = Error("unrecognized frame returned, CONNECT")
	EBADFRM = Error("Malformed frame")
	EUNKHDR = Error("currupt frame headers")

	// No body allowed error
	EBDYDATA = Error("body data not allowed")

	// Not connected.
	ECONBAD = Error("no current connection")

	// Destination required
	EREQDSTSND = Error("destination required, SEND")
	EREQDSTSUB = Error("destination required, SUBSCRIBE")
	EREQDOIUNS = Error("destination or id required, UNSUBSCRIBE")

	// Message ID required.
	EREQMIDACK = Error("message-id required, ACK") // 1.0, 1.1
	EREQIDACK  = Error("id required, ACK")         // 1.2

	// Subscription required (STOMP 1.1).
	EREQSUBACK = Error("subscription required, ACK")

	// NACK's.  STOMP 1.1 or greater.
	EREQMIDNAK = Error("message-id required, NACK")   // 1.1
	EREQSUBNAK = Error("subscription required, NACK") // 1.1
	EREQIDNAK  = Error("id required, NACK")           // 1.2

	// Transaction ID required.
	EREQTIDBEG = Error("transaction-id required, BEGIN")
	EREQTIDCOM = Error("transaction-id required, COMMIT")
	EREQTIDABT = Error("transaction-id required, ABORT")

	// Subscription errors.
	EDUPSID = Error("duplicate subscription-id")
	EBADSID = Error("invalid subscription-id")

	// Unscubscribe error.
	EUNOSID = Error("id required, UNSUBSCRIBE")

	// Unsupported version error.
	EBADVERCLI = Error("unsupported protocol version, client")
	EBADVERSVR = Error("unsupported protocol version, server")
	EBADVERNAK = Error("unsupported protocol version, NACK")

	// Unsupported Headers type.
	EBADHDR = Error("unsupported Headers type")

	// Receipt not allowed on connect
	ENORECPT = Error("receipt not allowed on CONNECT")
)

/*
	A zero length buffer for convenience.
*/
var NULLBUFF = make([]uint8, 0)

/*
	Codec data structure definition.
*/
type codecdata struct {
	encoded string
	decoded string
}

/*
	STOMP specification defined encoded / decoded values for the Message
	command and headers.
*/
var codec_values = []codecdata{
	codecdata{"\\\\", "\\"},
	codecdata{"\\" + "n", "\n"},
	codecdata{"\\" + "r", "\r"},
	codecdata{"\\c", ":"},
}

/*
	Control data for initialization of heartbeats with STOMP 1.1+, and the
	subsequent control of any heartbeat routines.
*/
type heartbeat_data struct {
	cx int64 // client send value, ms
	cy int64 // client receive value, ms
	sx int64 // server send value, ms
	sy int64 // server receive value, ms
	//
	hbs bool // sending heartbeats
	hbr bool // receiving heartbeats
	//
	sti int64 // local sender ticker interval, ns
	rti int64 // local receiver ticker interval, ns
	//
	sc int64 // local sender ticker count
	rc int64 // local receiver ticker count
	//
	ssd chan bool // sender shutdown channel
	rsd chan bool // receiver shutdown channel
	//
	ls int64 // last send time, ns
	lr int64 // last receive time, ns
}

/*
	Control structure for basic client metrics.
*/
type metrics struct {
	st  time.Time // Start Time
	tfr int64     // Total frame reads
	tbr int64     // Total bytes read
	tfw int64     // Total frame writes
	tbw int64     // Total bytes written
}
