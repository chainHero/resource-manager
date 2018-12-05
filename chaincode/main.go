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
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// ResourceManagerChaincode implementation of Chaincode
type ResourceManagerChaincode struct {
}

// Init of the chaincode
// This function is called only one when the chaincode is instantiated.
// So the goal is to prepare the ledger to handle future requests.
func (t *ResourceManagerChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### ResourceManagerChaincode Init ###########")

	// Get the function and arguments from the request
	function, _ := stub.GetFunctionAndParameters()

	// Check if the request is the init function
	if function != "init" {
		return shim.Error("Unknown function call")
	}

	// Return a successful message
	return shim.Success(nil)
}

// Invoke on the chaincode
// All future requests named invoke will arrive here.
// It is the main entry point for the chaincode after the init proceed.
func (t *ResourceManagerChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("########### ResourceManagerChaincode Invoke ###########")

	// Get the function and arguments from the request
	function, args := stub.GetFunctionAndParameters()

	// Check whether it is an invoke request
	if function != "invoke" {
		return shim.Error("Unknown function call")
	}

	// Check whether the number of arguments is sufficient
	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	// In order to manage multiple type of request, we will check the first argument.

	// Here we have one possible argument: query (every query request will read in the ledger without modification)
	if args[0] == "query" {
		return t.query(stub, args[1:])
	}

	// The update argument will manage all update in the ledger
	if args[0] == "update" {
		return t.update(stub, args[1:])
	}

	// If the arguments given don’t match any function, we return an error
	return shim.Error("Unknown action, check the first argument")
}

func main() {
	// Start the chaincode and make it ready for futures requests
	err := shim.Start(new(ResourceManagerChaincode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
