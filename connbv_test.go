//
// Copyright © 2012-2019 Guy M. Allard
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

package stompws

import "testing"

/*
	ConnBadValVer Test: Bad Version value.
*/
func TestConnBadValVer(t *testing.T) {
	for _, p := range Protocols() {
		n, _ = openConn(t)
		ch := login_headers
		ch = ch.Add(HK_ACCEPT_VERSION, "3.14159").Add(HK_HOST, "localhost")
		conn, e = Connect(n, ch)
		if e == nil {
			t.Errorf("TestConnBadValVer Expected error, got nil, proto: %s\n", p)
		}
		if e != EBADVERCLI {
			t.Errorf("TestConnBadValVer Expected <%v>, got <%v>, proto: %s\n",
				EBADVERCLI, e, p)
		}
		checkReceived(t, conn, false)
		// We are not connected by test design, check nothing around
		// DISCONNECT.
		_ = closeConn(t, n)
	}
}

/*
	ConnBadValHost Test: Bad Version, no host (vhost) value.
*/
func TestConnBadValHost(t *testing.T) {
	for _, p := range Protocols() {
		n, _ = openConn(t)
		ch := login_headers
		ch = ch.Add(HK_ACCEPT_VERSION, p)
		conn, e = Connect(n, ch)
		if e == nil {
			t.Errorf("TestConnBadValHost Expected error, got nil, proto: %s\n", p)
		}
		if e != EREQHOST {
			t.Errorf("TestConnBadValHost Expected <%v>, got <%v>, proto: %s\n",
				EREQHOST, e, p)
		}
		checkReceived(t, conn, false)
		// We are not connected by test design, check nothing around
		// DISCONNECT.
		_ = closeConn(t, n)
	}
}
