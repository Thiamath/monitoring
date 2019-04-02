/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package fake

import (
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
)

type FakeResult string

func (r FakeResult) ResultType() query.QueryProviderType {
	return FakeProviderType
}
