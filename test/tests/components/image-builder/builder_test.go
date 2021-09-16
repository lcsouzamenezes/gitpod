// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package imagebuilder

import (
	"context"
	"io"
	"testing"

	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	csapi "github.com/gitpod-io/gitpod/content-service/api"
	imgapi "github.com/gitpod-io/gitpod/image-builder/api"
	"github.com/gitpod-io/gitpod/test/pkg/integration"
	test_context "github.com/gitpod-io/gitpod/test/pkg/integration/context"
)

func TestBaseImageBuild(t *testing.T) {
	builtinUser := features.New("database").
		WithLabel("component", "image-builder").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it should exists a builtin user workspace", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			client, err := api.ImageBuilder(integration.SelectImageBuilderMK3)
			if err != nil {
				t.Fatal("cannot start build", err)
			}

			bld, err := client.Build(ctx, &imgapi.BuildRequest{
				ForceRebuild: true,
				Source: &imgapi.BuildSource{
					From: &imgapi.BuildSource_File{
						File: &imgapi.BuildSourceDockerfile{
							DockerfileVersion: "some-version",
							DockerfilePath:    ".gitpod.Dockerfile",
							ContextPath:       ".",
							Source: &csapi.WorkspaceInitializer{
								Spec: &csapi.WorkspaceInitializer_Git{
									Git: &csapi.GitInitializer{
										RemoteUri:  "https://github.com/gitpod-io/dazzle.git",
										TargetMode: csapi.CloneTargetMode_REMOTE_BRANCH,
										CloneTaget: "main",
										Config: &csapi.GitConfig{
											Authentication: csapi.GitAuthMethod_NO_AUTH,
										},
									},
								},
							},
						},
					},
				},
			})
			if err != nil {
				t.Fatal("cannot start build", err)
			}

			var ref string
			for {
				msg, err := bld.Recv()
				if err != nil && err != io.EOF {
					t.Fatal(err)
				}

				ref = msg.Ref
				if msg.Status == imgapi.BuildStatus_done_success {
					break
				} else if msg.Status == imgapi.BuildStatus_done_failure {
					t.Fatalf("image build failed: %s", msg.Message)
				} else {
					t.Logf("build output: %s", msg.Message)
				}
			}
			if ref == "" {
				t.Fatal("ref was empty")
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)
			defer api.Done(t)

			return ctx
		}).
		Feature()

	testEnv.Test(t, builtinUser)
}

func TestParallelBaseImageBuild(t *testing.T) {
	t.Skip("Skipping...")

	parallelBuild := features.New("image-builder").
		WithLabel("component", "image-builder").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := integration.NewComponentAPI(ctx, cfg.Namespace(), cfg.Client())
			return test_context.SetComponentAPI(ctx, api)
		}).
		Assess("it should allow parallel build of images", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			api := test_context.GetComponentAPI(ctx)

			client, err := api.ImageBuilder(integration.SelectImageBuilderMK3)
			if err != nil {
				t.Fatalf("cannot start build: %q", err)
			}

			req := &imgapi.BuildRequest{
				ForceRebuild: true,
				Source: &imgapi.BuildSource{
					From: &imgapi.BuildSource_File{
						File: &imgapi.BuildSourceDockerfile{
							DockerfileVersion: "some-version",
							DockerfilePath:    ".gitpod.Dockerfile",
							ContextPath:       ".",
							Source: &csapi.WorkspaceInitializer{
								Spec: &csapi.WorkspaceInitializer_Git{
									Git: &csapi.GitInitializer{
										RemoteUri:  "https://github.com/gitpod-io/dazzle.git",
										TargetMode: csapi.CloneTargetMode_REMOTE_BRANCH,
										CloneTaget: "main",
										Config: &csapi.GitConfig{
											Authentication: csapi.GitAuthMethod_NO_AUTH,
										},
									},
								},
							},
						},
					},
				},
			}

			bld0, err := client.Build(ctx, req)
			if err != nil {
				t.Fatalf("cannot start build: %v", err)
			}
			bld1, err := client.Build(ctx, req)
			if err != nil {
				t.Fatalf("cannot start build: %v", err)
			}

			watchBuild := func(cl imgapi.ImageBuilder_BuildClient) error {
				for {
					msg, err := cl.Recv()
					if err != nil && err != io.EOF {
						return xerrors.Errorf("image builder error: %v", err)
					}
					if err := ctx.Err(); err != nil {
						return xerrors.Errorf("context error: %v", err)
					}

					if msg.Status == imgapi.BuildStatus_done_success {
						break
					} else if msg.Status == imgapi.BuildStatus_done_failure {
						return xerrors.Errorf("image build failed: %s", msg.Message)
					} else {
						t.Logf("build output: %s", msg.Message)
					}
				}

				return nil
			}

			var eg errgroup.Group
			eg.Go(func() error { return watchBuild(bld0) })
			eg.Go(func() error { return watchBuild(bld1) })

			blds, err := client.ListBuilds(ctx, &imgapi.ListBuildsRequest{})
			if err != nil {
				t.Fatalf("cannot list builds: %v", err)
			}

			// TODO(cw): make this assertion resiliant against other on-going builds
			if l := len(blds.Builds); l != 1 {
				t.Errorf("image builder is running not just one build, but %d", l)
			}

			err = eg.Wait()
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

	testEnv.Test(t, parallelBuild)
}
