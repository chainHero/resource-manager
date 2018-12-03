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
	"encoding/base64"
	"fmt"
	"github.com/chainHero/resource-manager/app/fabric"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Constants used to know if a form is submitted
const (
	formSubmittedKey   = "submitted"
	formSubmittedValue = "true"
)

// Controller struct use to store a Fabric SDK instance and serve web pages
type Controller struct {
	Fabric *fabric.Setup
}

// basicAuth used to check the authentication (using basic auth) and retrieve the blockchain user
func (c *Controller) basicAuth(pass func(http.ResponseWriter, *http.Request, *fabric.User)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		u, err := c.Fabric.LogUser(pair[0], pair[1])
		if err != nil {
			http.Error(w, fmt.Sprintf("authorization failed with error: %v", err), http.StatusUnauthorized)
			return
		}

		pass(w, r, u)
	}
}

// LogoutHandler handler to disconnect the user (using basic auth)
func (c *Controller) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	w.WriteHeader(http.StatusUnauthorized)
	return
}

// renderTemplate allow to render a template using the predefined layout
func renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data interface{}) {
	lp := filepath.Join("web", "templates", "layout.gohtml")
	tp := filepath.Join("web", "templates", templateName)

	// Return a 404 if the template doesn't exist
	info, err := os.Stat(tp)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
	}

	// Return a 404 if the request is for a directory
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}

	resultTemplate, err := template.ParseFiles(tp, lp)
	if err != nil {
		// Log the detailed error
		fmt.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}
	if err := resultTemplate.ExecuteTemplate(w, "layout", data); err != nil {
		fmt.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
