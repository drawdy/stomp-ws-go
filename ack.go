//
// Copyright © 2011 Guy M. Allard
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

package stomp

import (
	"os"
)

// Ack
func (c *Connection) Ack(h Headers) (e os.Error) {
	if !c.connected {
		return ECONBAD
	}
	if _, ok := h.Contains("subscription"); !ok {
		return EREQSUBACK
	}
	if _, ok := h.Contains("message-id"); !ok {
		return EREQMIDACK
	}
	ch := h.Clone()
	return c.transmitCommon(ACK, ch)
}
