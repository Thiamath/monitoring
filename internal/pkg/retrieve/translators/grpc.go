/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// GRPC utils
// Translators from specific QueryResult implementations to the
// appropriate grpc.QueryResponse

package translators

import (
	"time"
	"github.com/golang/protobuf/ptypes/timestamp"
)


func GRPCTime(t time.Time) *timestamp.Timestamp {
	if t.IsZero() {
		return nil
	}
	return &timestamp.Timestamp{
		Seconds: t.Unix(),
		Nanos: int32(t.Nanosecond()),
	}
}
