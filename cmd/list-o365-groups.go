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

	"github.com/certmichelin/azurehound/v3/client"
	"github.com/certmichelin/azurehound/v3/client/query"
	"github.com/certmichelin/azurehound/v3/enums"
	"github.com/certmichelin/azurehound/v3/models"
	"github.com/certmichelin/azurehound/v3/panicrecovery"
	"github.com/certmichelin/azurehound/v3/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listO365GroupsCmd)
}

var listO365GroupsCmd = &cobra.Command{
	Use:          "O365Groups",
	Long:         "Lists Azure Active Directory Office 365 Groups",
	Run:          listO365GroupsCmdImpl,
	SilenceUsage: true,
}

func listO365GroupsCmdImpl(cmd *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure active directory Office 365 Groups...")
	start := time.Now()
	stream := listO365Groups(ctx, azClient)
	panicrecovery.HandleBubbledPanic(ctx, stop, log)
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listO365Groups(ctx context.Context, client client.AzureClient) <-chan interface{} {
	out := make(chan interface{})

	params := query.GraphParams{Select: []string{

		"authenticationType",
		"availabilityStatus",
		"id",
		"isAdminManaged",
		"isDefault",
		"isInitial",
		"isRoot",
		"isVerified",
		"passwordValidityPeriodInDays",
		"passwordNotificationWindowInDays",
		"state",
		// "tags",
	}}

	go func() {
		defer panicrecovery.PanicRecovery()
		defer close(out)
		count := 0
		for item := range client.ListAzureADO365Groups(ctx, params) {
			if item.Error != nil {
				log.Error(item.Error, "unable to continue processing domain")
				return
			} else {
				log.V(2).Info("found domain", "O365Groups", item)
				count++
				domain := models.O365Group{
					O365Group:  item.Ok,
					TenantId:   client.TenantInfo().TenantId,
					TenantName: client.TenantInfo().DisplayName,
				}
				if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
					Kind: enums.KindAZO365Group,
					Data: domain,
				}); !ok {
					return
				}
			}
		}
		log.Info("finished listing all Office 365 Groups", "count", count)
	}()

	return out
}
