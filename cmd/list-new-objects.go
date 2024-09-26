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

	"github.com/bloodhoundad/azurehound/v2/client"
	"github.com/bloodhoundad/azurehound/v2/client/query"
	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/bloodhoundad/azurehound/v2/panicrecovery"
	"github.com/bloodhoundad/azurehound/v2/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listNewObjectsCmd)
}

var listNewObjectsCmd = &cobra.Command{
	Use:          "new-objects",
	Long:         "Lists Azure Active Directory New Objects",
	Run:          listNewObjectsCmdImpl,
	SilenceUsage: true,
}

func listNewObjectsCmdImpl(cmd *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure active directory new objects...")
	start := time.Now()
	stream := listNewObjects(ctx, azClient)
	panicrecovery.HandleBubbledPanic(ctx, stop, log)
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listNewObjects(ctx context.Context, client client.AzureClient) <-chan interface{} {
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
		for item := range client.ListAzureADNewObjects(ctx, params) {
			if item.Error != nil {
				log.Error(item.Error, "unable to continue processing new object")
				return
			} else {
				log.V(2).Info("found new object", "new-object", item)
				count++
				newobject := models.NewObject{
					NewObject:  item.Ok,
					TenantId:   client.TenantInfo().TenantId,
					TenantName: client.TenantInfo().DisplayName,
				}
				if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
					Kind: enums.KindAZNewObject,
					Data: newobject,
				}); !ok {
					return
				}
			}
		}
		log.Info("finished listing all new objects", "count", count)
	}()

	return out
}
