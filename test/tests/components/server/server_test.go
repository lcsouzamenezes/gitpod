// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package server

import (
	"context"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	protocol "github.com/gitpod-io/gitpod/gitpod-protocol"
	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
)

func TestServerAccess(t *testing.T) {
	getLoggedInUser := features.New("GetLoggedInUser").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can get a not built-in logged user", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			defer api.Done(t)

			server, err := api.GitpodServer()
			if err != nil {
				t.Fatalf("cannot get GitpodServer: %q", err)
			}

			_, err = server.GetLoggedInUser(ctx)
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, getLoggedInUser)
}

func TestStartWorkspace(t *testing.T) {
	startWorkspace := features.New("CreateWorkspace").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can run workspace tasks", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			server, err := api.GitpodServer()
			if err != nil {
				t.Fatalf("cannot get GitpodServer: %q", err)
			}

			resp, err := server.CreateWorkspace(ctx, &protocol.CreateWorkspaceOptions{
				ContextURL: "github.com/gitpod-io/gitpod",
				Mode:       "force-new",
			})
			if err != nil {
				t.Fatalf("cannot start workspace: %q", err)
			}

			t.Cleanup(func() {
				cctx, ccancel := context.WithTimeout(context.Background(), 10*time.Second)
				err := server.StopWorkspace(cctx, resp.CreatedWorkspaceID)
				ccancel()
				if err != nil {
					t.Logf("cannot stop workspace: %q", err)
				}
			})

			t.Logf("created workspace: workspaceID=%s url=%s", resp.CreatedWorkspaceID, resp.WorkspaceURL)

			nfo, err := server.GetWorkspace(ctx, resp.CreatedWorkspaceID)
			if err != nil {
				t.Fatalf("cannot get workspace: %q", err)
			}
			if nfo.LatestInstance == nil {
				t.Fatal("CreateWorkspace did not start the workspace")
			}

			_, err = integration.WaitForWorkspaceStart(ctx, nfo.LatestInstance.ID, api)
			if err != nil {
				t.Fatalf("cannot get workspace: %q", err)
			}

			t.Logf("workspace is running: instanceID=%s", nfo.LatestInstance.ID)
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testEnv.Test(t, startWorkspace)
}
