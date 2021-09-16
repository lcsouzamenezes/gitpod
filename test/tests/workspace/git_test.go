// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package workspace

import (
	"context"
	"net/rpc"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	agent "github.com/gitpod-io/gitpod/test/pkg/agent/workspace/api"
	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
	"github.com/gitpod-io/gitpod/test/tests/workspace/common"
)

//
type GitTest struct {
	Skip          bool
	Name          string
	ContextURL    string
	WorkspaceRoot string
	Action        GitFunc
}

type GitFunc func(rsa *rpc.Client, git common.GitClient, workspaceRoot string) error

func TestGitActions(t *testing.T) {
	tests := []GitTest{
		{
			Name:          "create, add and commit",
			ContextURL:    "github.com/gitpod-io/gitpod-test-repo/tree/integration-test/commit-and-push",
			WorkspaceRoot: "/workspace/gitpod-test-repo",
			Action: func(rsa *rpc.Client, git common.GitClient, workspaceRoot string) (err error) {

				var resp agent.ExecResponse
				err = rsa.Call("WorkspaceAgent.Exec", &agent.ExecRequest{
					Dir:     workspaceRoot,
					Command: "bash",
					Args: []string{
						"-c",
						"echo \"another test run...\" >> file_to_commit.txt",
					},
				}, &resp)
				if err != nil {
					return err
				}
				err = git.Add(workspaceRoot)
				if err != nil {
					return err
				}
				err = git.Commit(workspaceRoot, "automatic test commit", false)
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Skip:          true,
			Name:          "create, add and commit and PUSH",
			ContextURL:    "github.com/gitpod-io/gitpod-test-repo/tree/integration-test/commit-and-push",
			WorkspaceRoot: "/workspace/gitpod-test-repo",
			Action: func(rsa *rpc.Client, git common.GitClient, workspaceRoot string) (err error) {

				var resp agent.ExecResponse
				err = rsa.Call("WorkspaceAgent.Exec", &agent.ExecRequest{
					Dir:     workspaceRoot,
					Command: "bash",
					Args: []string{
						"-c",
						"echo \"another test run...\" >> file_to_commit.txt",
					},
				}, &resp)
				if err != nil {
					return err
				}
				err = git.Add(workspaceRoot)
				if err != nil {
					return err
				}
				err = git.Commit(workspaceRoot, "automatic test commit", false)
				if err != nil {
					return err
				}
				err = git.Push(workspaceRoot, false)
				if err != nil {
					return err
				}
				return nil
			},
		},
	}

	gitActions := features.New("GitActions").
		WithLabel("component", "server").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it can run git actions", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			for _, test := range tests {
				t.Run(test.ContextURL, func(t *testing.T) {
					if test.Skip {
						t.SkipNow()
					}

					nfo, stopWS, err := integration.LaunchWorkspaceFromContextURL(ctx, test.ContextURL, api)
					if err != nil {
						t.Fatal(err)
					}

					defer stopWS(false)

					_, err = integration.WaitForWorkspaceStart(ctx, nfo.LatestInstance.ID, api)
					if err != nil {
						t.Fatal(err)
					}

					rsa, closer, err := integration.Instrument(integration.ComponentWorkspace, "workspace", cfg.Namespace(), cfg.Client(), integration.WithInstanceID(nfo.LatestInstance.ID))
					if err != nil {
						t.Fatal(err)
					}
					defer rsa.Close()
					defer integration.RunCloser(t, closer)

					git := common.Git(rsa)
					err = test.Action(rsa, git, test.WorkspaceRoot)
					if err != nil {
						t.Fatal(err)
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

	testEnv.Test(t, gitActions)
}
