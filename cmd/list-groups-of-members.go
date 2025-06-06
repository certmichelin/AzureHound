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
	listRootCmd.AddCommand(listGroupsOfMemberCmd)
	listGroupsOfMemberCmd.Flags().StringVarP(&groupID, "group-id", "g", "", "ID du groupe cible")
	listGroupsOfMemberCmd.MarkFlagRequired("group-id")
	//listGroupsOfMemberCmd.Flags().StringSliceVar(&listGroupsOfMemberSelect, "select", []string{"id,displayName,createdDateTime"}, `Select properties to include. Use "" for Azure default properties. Azurehound default is "id,displayName,createdDateTime" if flag is not supplied.`)
}

var listGroupsOfMemberCmd = &cobra.Command{
	Use:          "groups-of-member",
	Long:         "Lists Azure AD Group Members",
	Run:          listGroupsOfMemberCmdImpl,
	SilenceUsage: true,
}

var listGroupsOfMemberSelect []string

func listGroupsOfMemberCmdImpl(cmd *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, os.Kill)
	defer gracefulShutdown(stop)

	log.V(1).Info("testing connections")
	azClient := connectAndCreateClient()
	log.Info("collecting azure group members...")
	start := time.Now()
	stream := listGroupsOfMember(ctx, azClient, listMembersOfGroup(ctx, azClient, groupID))
	outputStream(ctx, stream)
	duration := time.Since(start)
	log.Info("collection completed", "duration", duration.String())
}

func listGroupsOfMember(ctx context.Context, client client.AzureClient, users <-chan interface{}) <-chan interface{} {
	var (
		out     = make(chan interface{})
		ids     = make(chan string)
		streams = pipeline.Demux(ctx.Done(), ids, config.ColStreamCount.Value().(int))
		wg      sync.WaitGroup
		params  = query.GraphParams{
			Select: unique(listGroupsOfMemberSelect),
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
				log.Error(fmt.Errorf("failed user type assertion"), "unable to continue enumerating members groups", "result", result)
				return
			} else {
				if ok := pipeline.Send(ctx.Done(), ids, user.Id); !ok {
					return
				}
			}
		}
	}()

	wg.Add(len(streams))
	count := 0
	var mutex sync.Mutex
	seenGroups := make(map[string]string)

	for i := range streams {
		stream := streams[i]
		go func() {
			defer panicrecovery.PanicRecovery()
			defer wg.Done()
			for id := range stream {
				for item := range client.ListAzureADGroupsOfMembers(ctx, id, params) {
					if item.Error != nil {
						log.Error(item.Error, "unable to continue processing groups")
						return
					} else {
						groupID := item.Ok.Id
						mutex.Lock()
						if _, exists := seenGroups[groupID]; exists {
							mutex.Unlock()
							continue // Skip already seen groups
						}
						seenGroups[groupID] = ""
						mutex.Unlock()
						log.V(2).Info("found group", "group", item)
						count++
						group := models.Group{
							Group:      item.Ok,
							TenantId:   client.TenantInfo().TenantId,
							TenantName: client.TenantInfo().DisplayName,
						}
						if ok := pipeline.SendAny(ctx.Done(), out, AzureWrapper{
							Kind: enums.KindAZGroup,
							Data: group,
						}); !ok {
							return
						}
					}
				}
			}
			log.Info("finished listing all groups", "count", count)
		}()
	}

	go func() {
		wg.Wait()
		close(out)
		log.Info("finished listing groups for all members", "count", count)
	}()

	return out
}
