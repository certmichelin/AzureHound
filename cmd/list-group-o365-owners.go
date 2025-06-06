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
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/bloodhoundad/azurehound/v2/client"
	"github.com/bloodhoundad/azurehound/v2/client/query"
	"github.com/bloodhoundad/azurehound/v2/config"
	"github.com/bloodhoundad/azurehound/v2/enums"
	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/bloodhoundad/azurehound/v2/panicrecovery"
	"github.com/bloodhoundad/azurehound/v2/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listGroup365OwnersCmd)
}

var listGroup365OwnersCmd = &cobra.Command{
	Use:          "group365-owners",
	Long:         "Lists Azure AD Group Owners",
	Run:          listGroup365OwnersCmdImpl,
	SilenceUsage: true,
}

func listGroup365OwnersCmdImpl(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure group owners...")
	start := time.Now()
	stream := listGroup365Owners(ctx, azClient, listGroups365(ctx, azClient))
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listGroup365Owners(ctx context.Context, client client.AzureClient, groups <-chan interface{}) <-chan interface{} {
	var (
		out     = make(chan interface{})
		ids     = make(chan string)
		streams = pipeline.Demux(ctx.Done(), ids, config.ColStreamCount.Value().(int))
		wg      sync.WaitGroup
		params  = query.GraphParams{}
	)

	go func() {
		defer panicrecovery.PanicRecovery()
		defer close(ids)

		for result := range pipeline.OrDone(ctx.Done(), groups) {
			if group, ok := result.(AzureWrapper).Data.(models.Group365); !ok {
				log.Error(fmt.Errorf("failed type assertion"), "unable to continue enumerating group owners", "result", result)
				return
			} else {
				if ok := pipeline.Send(ctx.Done(), ids, group.Id); !ok {
					return
				}
			}
		}
	}()

	wg.Add(len(streams))
	for i := range streams {
		stream := streams[i]
		go func() {
			defer panicrecovery.PanicRecovery()
			defer wg.Done()
			for id := range stream {
				var (
					groupOwners = models.Group365Owners{
						GroupId: id,
					}
					count = 0
				)
				for item := range client.ListAzureADGroup365Owners(ctx, id, params) {
					if item.Error != nil {
						log.Error(item.Error, "unable to continue processing owners for this Microsoft 365 group", "groupId", id)
					} else {
						groupOwner := models.Group365Owner{
							Owner:   item.Ok,
							GroupId: id,
						}
						log.V(2).Info("found Microsoft 365 group owner", "groupOwner", groupOwner)
						count++
						groupOwners.Owners = append(groupOwners.Owners, groupOwner)
					}
				}
				if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
					Kind: enums.KindAZGroup365Owner,
					Data: groupOwners,
				}); !ok {
					return
				}
				log.V(1).Info("finished listing Microsoft 365 group owners", "groupId", id, "count", count)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		log.Info("finished listing all Microsoft 365 group owners")
	}()

	return out
}
