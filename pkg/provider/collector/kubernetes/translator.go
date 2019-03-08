/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates Kubernetes events to Platform counters

package kubernetes

import (
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// List of objects we are able to translate
var Translatable = []schema.GroupVersionKind{
	apps_v1.SchemeGroupVersion.WithKind("Deployment"),
	core_v1.SchemeGroupVersion.WithKind("Namespace"),
}
