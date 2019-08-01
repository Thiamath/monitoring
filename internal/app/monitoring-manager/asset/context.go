/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package asset

import (
	"context"
	"time"
)

const InventoryContextTimeout = 30 * time.Second
const ProxyContextTimeout = 30 * time.Second

// InventoryContext generates a new gRPC context for inventory connections
func InventoryContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), InventoryContextTimeout)
}

// ProxyContext generates a new gRPC context for edge inventory proxy connections
func ProxyContext() (context.Context, func()) {
	return context.WithTimeout(context.Background(), ProxyContextTimeout)
}
