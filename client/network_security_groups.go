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

	"github.com/certmichelin/azurehound/v3/client/query"
	"github.com/certmichelin/azurehound/v3/models/azure"
)

// List NSGs : https://learn.microsoft.com/en-us/rest/api/virtualnetwork/network-security-groups
func (s *azureClient) ListAzureNetworkSecurityGroups(ctx context.Context, subscriptionId string) <-chan AzureResult[azure.NetworkSecurityGroup] {
	var (
		out    = make(chan AzureResult[azure.NetworkSecurityGroup])
		path   = fmt.Sprintf("/subscriptions/%s/providers/Microsoft.Network/networkSecurityGroups", subscriptionId)
		params = query.RMParams{ApiVersion: "2024-03-01"}
	)

	log.Printf("Listing network security groups : %s", path)
	go getAzureObjectList[azure.NetworkSecurityGroup](s.resourceManager, ctx, path, params, out)

	return out
}
