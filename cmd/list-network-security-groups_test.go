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
	"testing"

	"github.com/certmichelin/azurehound/v3/client"
	"github.com/certmichelin/azurehound/v3/client/mocks"
	"github.com/certmichelin/azurehound/v3/models"
	"github.com/certmichelin/azurehound/v3/models/azure"
	"go.uber.org/mock/gomock"
)

func init() {
	setupLogger()
}

func TestListNetworkSecurityGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ctx := context.Background()

	mockClient := mocks.NewMockAzureClient(ctrl)

	mockSubscriptionsChannel := make(chan interface{})
	mockNetworkSecurityGroupChannel := make(chan client.AzureResult[azure.NetworkSecurityGroup])
	mockNetworkSecurityGroupChannel2 := make(chan client.AzureResult[azure.NetworkSecurityGroup])

	mockTenant := azure.Tenant{}
	mockError := fmt.Errorf("I'm an error")
	mockClient.EXPECT().TenantInfo().Return(mockTenant).AnyTimes()
	mockClient.EXPECT().ListAzureNetworkSecurityGroups(gomock.Any(), gomock.Any()).Return(mockNetworkSecurityGroupChannel).Times(1)
	mockClient.EXPECT().ListAzureNetworkSecurityGroups(gomock.Any(), gomock.Any()).Return(mockNetworkSecurityGroupChannel2).Times(1)
	channel := listNetworkSecurityGroups(ctx, mockClient, mockSubscriptionsChannel)

	go func() {
		defer close(mockSubscriptionsChannel)
		mockSubscriptionsChannel <- AzureWrapper{
			Data: models.Subscription{},
		}
		mockSubscriptionsChannel <- AzureWrapper{
			Data: models.Subscription{},
		}
	}()
	go func() {
		defer close(mockNetworkSecurityGroupChannel)
		mockNetworkSecurityGroupChannel <- client.AzureResult[azure.NetworkSecurityGroup]{
			Ok: azure.NetworkSecurityGroup{},
		}
		mockNetworkSecurityGroupChannel <- client.AzureResult[azure.NetworkSecurityGroup]{
			Ok: azure.NetworkSecurityGroup{},
		}
	}()
	go func() {
		defer close(mockNetworkSecurityGroupChannel2)
		mockNetworkSecurityGroupChannel2 <- client.AzureResult[azure.NetworkSecurityGroup]{
			Ok: azure.NetworkSecurityGroup{},
		}
		mockNetworkSecurityGroupChannel2 <- client.AzureResult[azure.NetworkSecurityGroup]{
			Error: mockError,
		}
	}()

	if result, ok := <-channel; !ok {
		t.Fatalf("failed to receive from channel")
	} else if wrapper, ok := result.(AzureWrapper); !ok {
		t.Errorf("failed type assertion: got %T, want %T", result, AzureWrapper{})
	} else if _, ok := wrapper.Data.(models.NetworkSecurityGroup); !ok {
		t.Errorf("failed type assertion: got %T, want %T", wrapper.Data, models.NetworkSecurityGroup{})
	}

	if result, ok := <-channel; !ok {
		t.Fatalf("failed to receive from channel")
	} else if wrapper, ok := result.(AzureWrapper); !ok {
		t.Errorf("failed type assertion: got %T, want %T", result, AzureWrapper{})
	} else if _, ok := wrapper.Data.(models.NetworkSecurityGroup); !ok {
		t.Errorf("failed type assertion: got %T, want %T", wrapper.Data, models.NetworkSecurityGroup{})
	}

	if result, ok := <-channel; !ok {
		t.Fatalf("failed to receive from channel")
	} else if wrapper, ok := result.(AzureWrapper); !ok {
		t.Errorf("failed type assertion: got %T, want %T", result, AzureWrapper{})
	} else if _, ok := wrapper.Data.(models.NetworkSecurityGroup); !ok {
		t.Errorf("failed type assertion: got %T, want %T", wrapper.Data, models.NetworkSecurityGroup{})
	}

	if _, ok := <-channel; ok {
		t.Error("should not have recieved from channel")
	}
}