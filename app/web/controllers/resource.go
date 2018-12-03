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

// ResourceHandler controller that allow to see resource details
func (c *Controller) ResourceHandler() func(http.ResponseWriter, *http.Request) {
	return c.basicAuth(func(w http.ResponseWriter, r *http.Request, u *fabric.User) {

		// Check that the user connected is an admin, else return to the home page
		_, err := u.QueryAdmin()
		if err != nil {
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
			return
		}

		resourceID := r.URL.Query().Get("id")
		if resourceID == "" {
			http.Redirect(w, r, "/resources", http.StatusTemporaryRedirect)
			return
		}

		resource, resourcesHistory, err := u.QueryResource(resourceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to retrieve resource detail from the ledger: %v", err), http.StatusInternalServerError)
			return
		}

		data := &struct {
			Username  string
			Resource  *model.Resource
			Histories model.ResourceHistories
			IsDeleted bool
		}{
			Username:  u.Username,
			Resource:  resource,
			Histories: resourcesHistory,
			IsDeleted: len(resourcesHistory) > 0 && resourcesHistory[0].Deleted,
		}
		renderTemplate(w, r, "resource.gohtml", data)
	})
}
