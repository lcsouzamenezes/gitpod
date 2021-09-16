// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package wsmanager

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
	wsmanager_api "github.com/gitpod-io/gitpod/ws-manager/api"
)

func TestGetWorkspaces(t *testing.T) {
	getWorkspaces := features.New("workspaces").
		WithLabel("component", "ws-manager").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it should get workspaces", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			wsman, err := api.WorkspaceManager()
			if err != nil {
				t.Fatal(err)
			}

			_, err = wsman.GetWorkspaces(ctx, &wsmanager_api.GetWorkspacesRequest{})
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testEnv.Test(t, getWorkspaces)
}
