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

	"github.com/certmichelin/azurehound/v3/client"
	"github.com/certmichelin/azurehound/v3/config"
	"github.com/certmichelin/azurehound/v3/enums"
	"github.com/certmichelin/azurehound/v3/models"
	"github.com/certmichelin/azurehound/v3/panicrecovery"
	"github.com/certmichelin/azurehound/v3/pipeline"
	"github.com/spf13/cobra"
)

func init() {
	listRootCmd.AddCommand(listVMScaleSetsCmd)
}

var listVMScaleSetsCmd = &cobra.Command{
	Use:          "vm-scale-sets",
	Long:         "Lists Azure Virtual Machine Scale Sets",
	Run:          listVMScaleSetsCmdImpl,
	SilenceUsage: true,
}

func listVMScaleSetsCmdImpl(cmd *cobra.Command, args []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	if err := testConnections(); err != nil {
		exit(err)
	} else if azClient, err := newAzureClient(); err != nil {
		exit(err)
	} else {
		log.Info("collecting azure virtual machine scale sets...")
		start := time.Now()
		stream := listVMScaleSets(ctx, azClient, listSubscriptions(ctx, azClient))
		panicrecovery.HandleBubbledPanic(ctx, stop, log)
		outputStream(ctx, stream)
		duration := time.Since(start)
		log.Info("collection completed", "duration", duration.String())
	}

}

func listVMScaleSets(ctx context.Context, client client.AzureClient, subscriptions <-chan interface{}) <-chan interface{} {
	var (
		out     = make(chan interface{})
		ids     = make(chan string)
		streams = pipeline.Demux(ctx.Done(), ids, config.ColStreamCount.Value().(int))
		wg      sync.WaitGroup
	)

	go func() {
		defer panicrecovery.PanicRecovery()
		defer close(ids)
		for result := range pipeline.OrDone(ctx.Done(), subscriptions) {
			if subscription, ok := result.(AzureWrapper).Data.(models.Subscription); !ok {
				log.Error(fmt.Errorf("failed type assertion"), "unable to continue enumerating virtual machine scale sets", "result", result)
				return
			} else {
				if ok := pipeline.Send(ctx.Done(), ids, subscription.SubscriptionId); !ok {
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
				count := 0
				for item := range client.ListAzureVMScaleSets(ctx, id) {
					if item.Error != nil {
						log.Error(item.Error, "unable to continue processing virtual machine scale sets for this subscription", "subscriptionId", id)
					} else {
						vmScaleSet := models.VMScaleSet{
							VMScaleSet:      item.Ok,
							SubscriptionId:  "/subscriptions/" + id,
							ResourceGroupId: item.Ok.ResourceGroupId(),
							TenantId:        client.TenantInfo().TenantId,
						}
						log.V(2).Info("found virtual machine scale set", "vmScaleSet", vmScaleSet)
						count++
						if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
							Kind: enums.KindAZVMScaleSet,
							Data: vmScaleSet,
						}); !ok {
							return
						}
					}
				}
				log.V(1).Info("finished listing virtual machine scale sets", "subscriptionId", id, "count", count)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		log.Info("finished listing all virtual machine scale sets")
	}()

	return out
}
