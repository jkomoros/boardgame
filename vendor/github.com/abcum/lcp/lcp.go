// Copyright © 2016 Abcum Ltd
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

package lcp

import "bytes"

// LCP returns the longest common prefix between multiple slices of bytes.
// If no items are passed in to the method, then a nil byte slice is
// returned. If two or more byte slices are passed in to the method, then
// the longest common lexicographical prefix shared by all of the slices
// will be returned.
func LCP(items ...[]byte) []byte {

	switch len(items) {
	case 0:
		return nil
	case 1:
		return items[0]
	}

	min, max := items[0], items[0]
	for _, item := range items[1:] {
		switch {
		case bytes.Compare(item, min) == -1:
			min = item
		case bytes.Compare(item, max) == +1:
			max = item
		}
	}

	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			if i == 0 {
				return nil
			}
			return min[:i]
		}
	}

	if len(min) == 0 {
		return nil
	}

	return min

}
