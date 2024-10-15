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

package azure

type NetworkSecurityGroupProperties struct {
	ProvisioningState    string                                  `json:"provisioningState,omitempty"`
	ResourceGuid         string                                  `json:"resourceGuid,omitempty"`
	SecurityRules        []NetworkSecurityGroupSecurityRules     `json:"securityRules,omitempty"`
	DefaultSecurityRules []NetworkSecurityGroupSecurityRules     `json:"defaultSecurityRules,omitempty"`
	NetworkInterfaces    []NetworkSecurityGroupNetworkInterfaces `json:"networkInterfaces,omitempty"`
}

type NetworkSecurityGroupSecurityRules struct {
	SRName       string                                     `json:"name,omitempty"`
	SRId         string                                     `json:"id,omitempty"`
	SREtag       string                                     `json:"etag,omitempty"`
	SRType       string                                     `json:"type,omitempty"`
	SRProperties NetworkSecurityGroupSecurityRuleProperties `json:"properties,omitempty"`
}

type NetworkSecurityGroupSecurityRuleProperties struct {
	SRProvisioningState          string   `json:"provisioningState,omitempty"`
	SRProtocol                   string   `json:"protocol,omitempty"`
	SRSourcePortRange            string   `json:"sourcePortRange,omitempty"`
	SRSourceAddressPrefix        string   `json:"sourceAddressPrefix,omitempty"`
	SRDestinationAddressPrefix   string   `json:"destinationAddressPrefix,omitempty"`
	SRAccess                     string   `json:"access,omitempty"`
	SRPriority                   int      `json:"priority,omitempty"`
	SRDirection                  string   `json:"direction,omitempty"`
	SRSourcePortRanges           []string `json:"sourcePortRanges,omitempty"`
	SRDestinationPortRanges      []string `json:"destinationPortRanges,omitempty"`
	SRSourceAddressPrefixes      []string `json:"sourceAddressPrefixes,omitempty"`
	SRDestinationAddressPrefixes []string `json:"destinationAddressPrefixes,omitempty"`
}

type NetworkSecurityGroupNetworkInterfaces struct {
	Id string `json:"id,omitempty"`
}
