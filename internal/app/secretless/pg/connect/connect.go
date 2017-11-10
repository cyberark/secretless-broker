/*
Copyright 2017 Crunchy Data Solutions, Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package connect

import (
	"net"
)

func Send(connection net.Conn, message []byte) (int, error) {
	return connection.Write(message)
}

func Receive(connection net.Conn) ([]byte, int, error) {
	buffer := make([]byte, 4096)
	length, err := connection.Read(buffer)
	return buffer, length, err
}

func Connect(host string) (net.Conn, error) {
	connection, err := net.Dial("tcp", host)

	if err != nil {
		return nil, err
	}

	return connection, nil
}
