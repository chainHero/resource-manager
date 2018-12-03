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

package main

import (
	"flag"
	"fmt"
	"github.com/chainHero/resource-manager/app/fabric"
	"github.com/chainHero/resource-manager/app/web"
	"github.com/chainHero/resource-manager/app/web/controllers"
	"github.com/chainHero/resource-manager/chaincode/model"
	"os"
)

// Flags allow to run some parts of the code when executing binary
const (
	flagInstall  = "install"
	flagRegister = "register"
)

func main() {
	// Manage flags
	flagsParams := make(map[string]*bool)
	flagsParams[flagInstall] = flag.Bool(flagInstall, false, "If set, the Fabric channel will be create and the chaincode installed/instantiated.")
	flagsParams[flagRegister] = flag.Bool(flagRegister, false, "If set, users will be registered in the Fabric CA.")
	flag.Parse()

	// Definition of the Fabric SDK properties
	fSetup := fabric.Setup{
		ChannelID:           "mychannel",
		ChannelConfig:       os.Getenv("GOPATH") + "/src/github.com/chainHero/resource-manager/fixtures/artifacts/mychannel/channel.tx",
		ChaincodeID:         "chainhero-resource-manager",
		ChaincodeGoPath:     os.Getenv("GOPATH"),
		ChaincodePath:       "github.com/chainHero/resource-manager/chaincode/",
		ChaincodeVersion:    "v1.0.0",
		OrgID:               "org1",
		OrgMspID:            "Org1MSP",
		OrgAdminUser:        "Admin",
		OrdererOrgID:        "ordererorg",
		OrdererOrgAdminUser: "Admin",
		OrdererID:           "orderer.hf.chainhero.io",
		CaID:                "ca.org1.hf.chainhero.io",
		ConfigFile:          "config.yaml",
	}

	// Initialization of the Fabric SDK from the previously set properties
	err := fSetup.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
		return
	}
	if *flagsParams[flagInstall] {
		err = fSetup.Install()
		if err != nil {
			fmt.Printf("Unable to install the Fabric channel and chaincode: %v\n", err)
			return
		}
	}
	// Close SDK
	defer fSetup.CloseSDK()

	// Register our users
	if *flagsParams[flagRegister] {
		err = fSetup.RegisterUser("admin1", "password", model.ActorAdmin)
		if err != nil {
			fmt.Printf("Unable to register the user 'admin1': %v\n", err)
			return
		}
		err = fSetup.RegisterUser("consumer1", "password", model.ActorConsumer)
		if err != nil {
			fmt.Printf("Unable to register the user 'consumer1': %v\n", err)
			return
		}
		err = fSetup.RegisterUser("consumer2", "password", model.ActorConsumer)
		if err != nil {
			fmt.Printf("Unable to register the user 'consumer2': %v\n", err)
			return
		}
	}

	// Launch the web application listening
	app := &controllers.Controller{
		Fabric: &fSetup,
	}
	err = web.Serve(app)
	if err != nil {
		fmt.Printf("Unable to start the server: %v", err)
	}
}
