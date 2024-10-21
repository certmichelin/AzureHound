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
	"github.com/certmichelin/azurehound/v3/constants"
	"github.com/certmichelin/azurehound/v3/models/azure"
)

// List groups : https://learn.microsoft.com/fr-fr/graph/api/group-list
func (s *azureClient) ListAzureADO365Groups(ctx context.Context, params query.GraphParams) <-chan AzureResult[azure.O365Group] {
	var (
		out  = make(chan AzureResult[azure.O365Group])
		path = fmt.Sprintf("/%s/groups?$filter=groupTypes/any(c:c+eq+'Unified')", constants.GraphApiVersion)
	)

	if params.Top == 0 {
		params.Top = 999
	}

	log.Printf("Listing O365 Groups : %s", path)
	go getAzureObjectList[azure.O365Group](s.msgraph, ctx, path, params, out)

	return out
}
