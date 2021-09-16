// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package wsmanager

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	content_service_api "github.com/gitpod-io/gitpod/content-service/api"
	agent "github.com/gitpod-io/gitpod/test/pkg/agent/workspace/api"
	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
	wsapi "github.com/gitpod-io/gitpod/ws-manager/api"
)

// TestBackup tests a basic start/modify/restart cycle
func TestBackup(t *testing.T) {
	backup := features.New("backup").
		WithLabel("component", "ws-manager").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it should start a workspace, create and file and successfully create a backup", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			api := test_context.GetComponentAPI(ctx)

			wsm, err := api.WorkspaceManager()
			if err != nil {
				t.Fatal(err)
			}

			ws, err := integration.LaunchWorkspaceDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			rsa, closer, err := integration.Instrument(integration.ComponentWorkspace, "workspace", cfg.Namespace(), cfg.Client(),
				integration.WithInstanceID(ws.Req.Id),
				integration.WithContainer("workspace"),
				integration.WithWorkspacekitLift(true),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer integration.RunCloser(t, closer)

			var resp agent.WriteFileResponse
			err = rsa.Call("WorkspaceAgent.WriteFile", &agent.WriteFileRequest{
				Path:    "/workspace/foobar.txt",
				Content: []byte("hello world"),
				Mode:    0644,
			}, &resp)
			if err != nil {
				_, _ = wsm.StopWorkspace(ctx, &wsapi.StopWorkspaceRequest{Id: ws.Req.Id})
				t.Fatal(err)
			}
			rsa.Close()

			_, err = wsm.StopWorkspace(ctx, &wsapi.StopWorkspaceRequest{
				Id: ws.Req.Id,
			})
			if err != nil {
				t.Fatal(err)
			}

			_, err = integration.WaitForWorkspaceStop(ctx, api, ws.Req.Id)
			if err != nil {
				t.Fatal(err)
			}

			ws, err = integration.LaunchWorkspaceDirectly(ctx, api,
				integration.WithRequestModifier(func(w *wsapi.StartWorkspaceRequest) error {
					w.ServicePrefix = ws.Req.ServicePrefix
					w.Metadata.MetaId = ws.Req.Metadata.MetaId
					w.Metadata.Owner = ws.Req.Metadata.Owner
					return nil
				}),
			)
			if err != nil {
				t.Fatal(err)
			}

			rsa, closer, err = integration.Instrument(integration.ComponentWorkspace, "workspace", cfg.Namespace(), cfg.Client(),
				integration.WithInstanceID(ws.Req.Id),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer integration.RunCloser(t, closer)

			var ls agent.ListDirResponse
			err = rsa.Call("WorkspaceAgent.ListDir", &agent.ListDirRequest{
				Dir: "/workspace",
			}, &ls)
			if err != nil {
				t.Fatal(err)
			}

			rsa.Close()

			var found bool
			for _, f := range ls.Files {
				if filepath.Base(f) == "foobar.txt" {
					found = true
					break
				}
			}
			if !found {
				t.Fatal("did not find foobar.txt from previous workspace instance")
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testEnv.Test(t, backup)

}

// TestMissingBackup ensures workspaces fail if they should have a backup but don't have one
func TestMissingBackup(t *testing.T) {
	startWorkspace := features.New("CreateWorkspace").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can run workspace tasks", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
			defer cancel()

			api := test_context.GetComponentAPI(ctx)

			ws, err := integration.LaunchWorkspaceDirectly(ctx, api)
			if err != nil {
				t.Fatal(err)
			}

			wsm, err := api.WorkspaceManager()
			if err != nil {
				t.Fatal(err)
			}

			_, err = wsm.StopWorkspace(ctx, &wsapi.StopWorkspaceRequest{Id: ws.Req.Id})
			if err != nil {
				t.Fatal(err)
			}

			_, err = integration.WaitForWorkspaceStop(ctx, api, ws.Req.Id)
			if err != nil {
				t.Fatal(err)
			}

			contentSvc, err := api.ContentService()
			if err != nil {
				t.Fatal(err)
			}

			_, err = contentSvc.DeleteWorkspace(ctx, &content_service_api.DeleteWorkspaceRequest{
				OwnerId:     ws.Req.Metadata.Owner,
				WorkspaceId: ws.Req.Metadata.MetaId,
			})
			if err != nil {
				t.Fatal(err)
			}

			tests := []struct {
				Name string
				FF   []wsapi.WorkspaceFeatureFlag
			}{
				{Name: "classic"},
				{Name: "fwb", FF: []wsapi.WorkspaceFeatureFlag{wsapi.WorkspaceFeatureFlag_FULL_WORKSPACE_BACKUP}},
			}
			for _, test := range tests {
				t.Run(test.Name+"_backup_init", func(t *testing.T) {
					testws, err := integration.LaunchWorkspaceDirectly(ctx, api, integration.WithRequestModifier(func(w *wsapi.StartWorkspaceRequest) error {
						w.ServicePrefix = ws.Req.ServicePrefix
						w.Metadata.MetaId = ws.Req.Metadata.MetaId
						w.Metadata.Owner = ws.Req.Metadata.Owner
						w.Spec.Initializer = &content_service_api.WorkspaceInitializer{
							Spec: &content_service_api.WorkspaceInitializer_Backup{
								Backup: &content_service_api.FromBackupInitializer{},
							},
						}
						w.Spec.FeatureFlags = test.FF
						return nil
					}), integration.WithWaitWorkspaceForOpts(integration.WorkspaceCanFail))
					if err != nil {
						t.Fatal(err)
					}

					if testws.LastStatus == nil {
						t.Fatal("did not receive a last status")
						return
					}
					if testws.LastStatus.Conditions.Failed == "" {
						t.Error("restarted workspace did not fail despite missing backup")
					}
				})
			}
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
