// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package workspace

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	agent "github.com/gitpod-io/gitpod/test/pkg/agent/workspace/api"
	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
)

func TestWorkspaceInstrumentation(t *testing.T) {
	instrumentation := features.New("instrumentation").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can instrument a workspace", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			nfo, stopWs, err := integration.LaunchWorkspaceFromContextURL(ctx, "github.com/gitpod-io/gitpod", api)
			if err != nil {
				t.Fatal(err)
			}
			defer stopWs(true)

			rsa, closer, err := integration.Instrument(integration.ComponentWorkspace, "workspace", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.LatestInstance.ID))
			if err != nil {
				t.Fatal(err)
			}
			defer rsa.Close()
			defer integration.RunCloser(t, closer)

			var ls agent.ListDirResponse
			err = rsa.Call("WorkspaceAgent.ListDir", &agent.ListDirRequest{
				Dir: "/workspace/gitpod",
			}, &ls)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range ls.Files {
				t.Log(f)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testEnv.Test(t, instrumentation)
}

func TestLaunchWorkspaceDirectly(t *testing.T) {
	launchWorkspace := features.New("workspace").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can run workspace tasks", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			nfo, err := integration.LaunchWorkspaceDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			err = integration.DeleteWorkspace(ctx, api, nfo.Req.Id)
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

	testEnv.Test(t, launchWorkspace)
}
