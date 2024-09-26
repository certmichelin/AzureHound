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

// Represents an Azure Active Directory user account.
type NewObject struct {
	DirectoryObject

	AuthenticationType               string `json:"authenticationType,omitempty"`
	AvailabilityStatus               int    `json:"availabilityStatus,omitempty"`
	Id                               string `json:"id,omitempty"`
	IsAdminManaged                   bool   `json:"isAdminManaged,omitempty"`
	IsDefault                        bool   `json:"isDefault,omitempty"`
	IsInitial                        bool   `json:"isInitial,omitempty"`
	IsRoot                           bool   `json:"isRoot,omitempty"`
	IsVerified                       bool   `json:"isVerified,omitempty"`
	PasswordValidityPeriodInDays     int    `json:"passwordValidityPeriodInDays,omitempty"`
	PasswordNotificationWindowInDays int    `json:"passwordNotificationWindowInDays,omitempty"`
	State                            int    `json:"state,omitempty"`
}
