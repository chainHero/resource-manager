// Copyright 2018 Antoine CHABERT, toHero.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"time"
)

// Actor metadata used for an admin and a consumer
type Actor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Available actor type
const (
	ActorAttribute = "actor"
	ActorConsumer  = "consumer"
	ActorAdmin     = "admin"
)

// Admin that manage resources available
type Admin struct {
	Actor
}

// Consumer that acquire and release some resources
type Consumer struct {
	Actor
}

// Resource that is manage by an admin actor and can be acquire and release by a consumer
type Resource struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
	Mission     string `json:"mission,omitempty"`
	Consumer    string `json:"consumer,omitempty"`
}

// ResourceHistory is a detailed information about a resource state in the ledger
type ResourceHistory struct {
	Transaction string    `json:"transaction"`
	Resource    Resource  `json:"value"`
	Time        time.Time `json:"time"`
	Deleted     bool      `json:"deleted"`
}

// ResourceHistories the list of state in the ledger of a resource (with sorting, older at the end)
type ResourceHistories []ResourceHistory

func (a ResourceHistories) Len() int           { return len(a) }
func (a ResourceHistories) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ResourceHistories) Less(i, j int) bool { return a[i].Time.After(a[j].Time) }

// ResourcesDeleted list of resources deleted
type ResourcesDeleted []Resource

// List of object type stored in the ledger
const (
	ObjectTypeAdmin            = "admin"
	ObjectTypeConsumer         = "consumer"
	ObjectTypeResource         = "resource"
	ObjectTypeResourcesDeleted = "resources-deleted"
)

// List of available filter for query resources
const (
	ResourcesFilterAll             = "all"
	ResourcesFilterOnlyAvailable   = "only-available"
	ResourcesFilterOnlyUnavailable = "only-unavailable"
)
