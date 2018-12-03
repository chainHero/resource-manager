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
	"net/http"
)

// AddResourceHandler controller that allow to add a resource
func (c *Controller) AddResourceHandler() func(http.ResponseWriter, *http.Request) {
	return c.basicAuth(func(w http.ResponseWriter, r *http.Request, u *fabric.User) {

		// Check that the user connected is an admin, else return to the home page
		_, err := u.QueryAdmin()
		if err != nil {
			http.Redirect(w, r, "/home", http.StatusTemporaryRedirect)
			return
		}

		data := &struct {
			Error    string
			Success  bool
			Response bool
			Username string
		}{
			Error:    "",
			Success:  false,
			Response: false,
			Username: u.Username,
		}
		if r.FormValue(formSubmittedKey) == formSubmittedValue {
			id := r.FormValue("id")
			description := r.FormValue("description")
			err := u.UpdateAdd(id, description)
			if err != nil {
				data.Error = fmt.Sprintf("Unable to make the transaction in the ledger: %v", err)
			} else {
				data.Success = true
			}
			data.Response = true
		}
		renderTemplate(w, r, "add-resource.gohtml", data)
	})
}
