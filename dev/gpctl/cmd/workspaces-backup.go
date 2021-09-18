// Copyright (c) 2020 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package cmd

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitpod-io/gitpod/common-go/log"
)

// workspacesBackupCmd represents the backup command
var workspacesBackupCmd = &cobra.Command{
	Use:   "backup <workspaceID>",
	Short: "backup a workspace",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		conn, client, err := getWorkspacesClient(ctx)
		if err != nil {
			log.WithError(err).Fatal("cannot connect")
		}
		defer conn.Close()

		instanceID := args[0]
		if strings.ContainsAny(instanceID, ".") || strings.HasPrefix(instanceID, "http://") || strings.HasPrefix(instanceID, "https://") {
			s, err := getStatusByURL(ctx, client, instanceID)
			if err != nil {
				log.Fatal(err)
			}
			instanceID = s.Id
		}

		log.WithField("instance id", instanceID).Debug("starting workspace backup")
		var resp interface{}
		// resp, err := client.BackupWorkspace(ctx, &api.BackupWorkspaceRequest{
		// 	Id:     instanceID,
		// 	Policy: policy,
		// })
		// if err != nil {
		// 	log.WithError(err).Fatal("error during RPC call")
		// }

		err = getOutputFormat("backing up\n", "").Print(resp)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	workspacesCmd.AddCommand(workspacesBackupCmd)
}
