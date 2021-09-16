// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package context

import (
	"context"

	"github.com/gitpod-io/gitpod/test/pkg/integration"
)

type componentKey struct{}

type workspaceKey struct{}

func GetComponentAPI(ctx context.Context) *integration.ComponentAPI {
	return ctx.Value(componentKey{}).(*integration.ComponentAPI)
}

func SetComponentAPI(ctx context.Context, api *integration.ComponentAPI) context.Context {
	return context.WithValue(ctx, componentKey{}, api)
}

func GetWorkspaceID(ctx context.Context) string {
	return ctx.Value(workspaceKey{}).(string)
}

func SetWorkspaceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, workspaceKey{}, id)
}
