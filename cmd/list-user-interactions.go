// Copyright (C) 2025 Specter Ops, Inc.
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
	listRootCmd.AddCommand(listUsersInteractionsCmd)
	listUsersInteractionsCmd.Flags().StringSliceVar(&listUsersInteractionsSelect, "select", []string{"id,displayName"}, `Select properties to include. Use "" for Azure default properties. Azurehound default is "id,displayName,createdDateTime" if flag is not supplied.`)
}

var listUsersInteractionsCmd = &cobra.Command{
	Use:          "users-interactions",
	Long:         "Lists people the user interact the most with",
	Run:          listUsersInteractionsCmdImpl,
	SilenceUsage: true,
}

var listUsersInteractionsSelect []string

func listUsersInteractionsCmdImpl(cmd *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting users interactions...")
	start := time.Now()
	stream := listUsersInteractions(ctx, azClient, listUsers(ctx, azClient))
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listUsersInteractions(ctx context.Context, client client.AzureClient, users <-chan interface{}) <-chan interface{} {
	var (
		out     = make(chan interface{})
		ids     = make(chan string)
		streams = pipeline.Demux(ctx.Done(), ids, config.ColStreamCount.Value().(int))
		wg      sync.WaitGroup
		params  = query.GraphParams{
			Select: unique(listUsersInteractionsSelect),
			Filter: "",
			Count:  false,
			Search: "",
			Top:    0,
			Expand: "",
		}
	)

	go func() {
		defer panicrecovery.PanicRecovery()
		defer close(ids)

		for result := range pipeline.OrDone(ctx.Done(), users) {
			if user, ok := result.(AzureWrapper).Data.(models.User); !ok {
				log.Error(fmt.Errorf("failed user type assertion"), "unable to continue enumerating user interactions", "result", result)
				return
			} else {
				if ok := pipeline.Send(ctx.Done(), ids, user.Id); !ok {
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
					data = models.UsersInteractions{
						UserId: id,
					}
					count = 0
				)
				for item := range client.ListAzureADUsersInteractions(ctx, id, params) {
					if item.Error != nil {
						log.Error(item.Error, "unable to continue processing users interactions for this user", "userId", id)
					} else {
						userinteraction := models.UserInteraction{
							User:   item.Ok,
							UserId: id,
						}
						log.V(2).Info("found interaction", "userinteraction", userinteraction)
						count++
						data.Users = append(data.Users, userinteraction)
					}
				}
				if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
					Kind: enums.KindAZUserInteraction,
					Data: data,
				}); !ok {
					return
				}
				log.V(1).Info("finished listing all user interactions", "userId", id, "count", count)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		log.Info("finished listing user interactions for all users")
	}()

	return out
}
