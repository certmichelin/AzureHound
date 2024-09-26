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

package client

import (
	"context"
	"fmt"
	"log"

	"github.com/bloodhoundad/azurehound/v2/client/query"
	"github.com/bloodhoundad/azurehound/v2/models/azure"
)

// ListDomains https://learn.microsoft.com/en-us/graph/api/user-findrooms
func (s *azureClient) ListAzureNewObjects(ctx context.Context, subscriptionId string, params query.RMParams) <-chan AzureResult[azure.NewObject] {
	var (
		out  = make(chan AzureResult[azure.NewObject])
		path = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Network/networkSecurityGroups", subscriptionId)
	)

	if params.ApiVersion == "" {
		params.ApiVersion = "2024-03-01"
	}

	log.Printf("Listing new objects : %s", path)
	go getAzureObjectList[azure.NewObject](s.msgraph, ctx, path, params, out)

	return out
}
