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

// ReleaseResourceHandler controller that allow to release a resource
func (c *Controller) ReleaseResourceHandler() func(http.ResponseWriter, *http.Request) {
	return c.basicAuth(func(w http.ResponseWriter, r *http.Request, u *fabric.User) {

		preSelectedResource := r.URL.Query().Get("id")

		data := &struct {
			Error               string
			Success             bool
			Response            bool
			PreSelectedResource string
			Resources           []model.Resource
			Username            string
		}{
			Error:               "",
			Success:             false,
			Response:            false,
			PreSelectedResource: preSelectedResource,
			Resources:           []model.Resource{},
			Username:            u.Username,
		}
		if r.FormValue(formSubmittedKey) == formSubmittedValue {

			// TODO implement the submit part in order to release a resource. Use 'id' provided in form parameter.
			// Tips:
			//  - read and understand the add and acquire form submit

			data.Response = true
			data.Error = "Unable to make the transaction in the ledger: not implemented in controller ReleaseResourceHandler"
		}

		resources, err := u.QueryResources(model.ResourcesFilterOnlyUnavailable)
		if err != nil {
			http.Error(w, fmt.Sprintf("Unable to retrieve resources from the ledger: %v", err), http.StatusInternalServerError)
			return
		}
		data.Resources = resources

		renderTemplate(w, r, "release-resource.gohtml", data)
	})
}
