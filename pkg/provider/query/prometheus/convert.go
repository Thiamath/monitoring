/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
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