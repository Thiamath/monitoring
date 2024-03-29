/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package prometheus

import (
	"fmt"
	"math"
)

func Ftoi(f float64) (int64, error) {
	if math.IsNaN(f) {
		return 0, nil
	}

	i := int64(f)
	if float64(i) != math.Trunc(f) {
		return 0, fmt.Errorf("float %f out of int64 range", f)
	}
	return i, nil
}
