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

	EmailAddress           string `json:"emailAddress,omitempty"`
	ID                     string `json:"id,omitempty"`
	DisplayName            string `json:"displayName,omitempty"`
	Address                string `json:"address,omitempty"`
	GeoCoordinates         string `json:"geoCoordinates,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	Nickname               string `json:"nickname,omitempty"`
	Building               string `json:"building,omitempty"`
	FloorNumber            string `json:"floorNumber,omitempty"`
	FloorLabel             string `json:"floorLabel,omitempty"`
	Label                  string `json:"label,omitempty"`
	Capacity               int    `json:"capacity,omitempty"`
	BookingType            string `json:"bookingType,omitempty"`
	AudioDeviceName        string `json:"audioDeviceName,omitempty"`
	VideoDeviceName        string `json:"videoDeviceName,omitempty"`
	DisplayDeviceName      string `json:"displayDeviceName,omitempty"`
	IsWheelChairAccessible bool   `json:"isWheelChairAccessible,omitempty"`
}
