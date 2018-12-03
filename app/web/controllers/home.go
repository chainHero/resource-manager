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

package controllers

import (
	"fmt"
	"github.com/chainHero/resource-manager/app/fabric"
	"github.com/chainHero/resource-manager/chaincode/model"
	"net/http"
)

// HomeHandler controller that allow to get the home page
func (c *Controller) HomeHandler() func(http.ResponseWriter, *http.Request) {
	return c.basicAuth(func(w http.ResponseWriter, r *http.Request, u *fabric.User) {

		resources, err := u.QueryResources(model.ResourcesFilterAll)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to retrieve resources from the ledger: %v", err), http.StatusInternalServerError)
			return
		}

		var resourcesCount uint64
		var resourcesAvailableCount uint64
		var resourcesUnavailableCount uint64

		for _, resource := range resources {
			resourcesCount++
			if resource.Available {
				resourcesAvailableCount++
			} else {
				resourcesUnavailableCount++
			}
		}

		data := &struct {
			Username                  string
			ResourcesCount            uint64
			ResourcesAvailableCount   uint64
			ResourcesUnavailableCount uint64
		}{
			Username:                  u.Username,
			ResourcesCount:            resourcesCount,
			ResourcesAvailableCount:   resourcesAvailableCount,
			ResourcesUnavailableCount: resourcesUnavailableCount,
		}
		renderTemplate(w, r, "home.gohtml", data)
	})
}
