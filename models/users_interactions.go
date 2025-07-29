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

package models

import (
	"encoding/json"
)

type UserInteraction struct {
	User   json.RawMessage `json:"user"`
	UserId string          `json:"UserId"`
}

func (s *UserInteraction) MarshalJSON() ([]byte, error) {
	output := make(map[string]any)
	output["userId"] = s.UserId

	if user, err := OmitEmpty(s.User); err != nil {
		return nil, err
	} else {
		output["user"] = user
		return json.Marshal(output)
	}
}

type UsersInteractions struct {
	Users  []UserInteraction `json:"users"`
	UserId string            `json:"userId"`
}
