// Copyright (C) 2022 Specter Ops, Inc.
//
// This file is part of AzureHound.
//
// AzureHound is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// AzureHound is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/certmichelin/azurehound/v3/constants"
	"github.com/certmichelin/azurehound/v3/enums"
	"github.com/certmichelin/azurehound/v3/internal"
	"github.com/certmichelin/azurehound/v3/models"
	"github.com/certmichelin/azurehound/v3/panicrecovery"
	"github.com/certmichelin/azurehound/v3/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listKeyVaultContributorsCmd)
}

var listKeyVaultContributorsCmd = &cobra.Command{
	Use:          "key-vault-contributors",
	Long:         "Lists Azure Key Vault Contributors",
	Run:          listKeyVaultContributorsCmdImpl,
	SilenceUsage: true,
}

func listKeyVaultContributorsCmdImpl(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure key vault contributors...")
	start := time.Now()
	subscriptions := listSubscriptions(ctx, azClient)
	keyVaults := listKeyVaults(ctx, azClient, subscriptions)
	kvRoleAssignments := listKeyVaultRoleAssignments(ctx, azClient, keyVaults)
	panicrecovery.HandleBubbledPanic(ctx, stop, log)
	stream := listKeyVaultContributors(ctx, kvRoleAssignments)
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listKeyVaultContributors(
	ctx context.Context,
	kvRoleAssignments <-chan azureWrapper[models.KeyVaultRoleAssignments],
) <-chan any {
	return pipeline.Map(ctx.Done(), kvRoleAssignments, func(ra azureWrapper[models.KeyVaultRoleAssignments]) any {
		filteredAssignments := internal.Filter(ra.Data.RoleAssignments, kvRoleAssignmentFilter(constants.ContributorRoleID))

		contributors := internal.Map(filteredAssignments, func(ra models.KeyVaultRoleAssignment) models.KeyVaultContributor {
			return models.KeyVaultContributor{
				Contributor: ra.RoleAssignment,
				KeyVaultId:  ra.KeyVaultId,
			}
		})

		return NewAzureWrapper(enums.KindAZKeyVaultContributor, models.KeyVaultContributors{
			KeyVaultId:   ra.Data.KeyVaultId,
			Contributors: contributors,
		})
	})
}
