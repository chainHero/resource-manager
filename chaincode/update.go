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
	"github.com/chainHero/resource-manager/chaincode/model"
	"github.com/hyperledger/fabric/core/chaincode/lib/cid"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

// update that handle every write in the ledger
func (t *ResourceManagerChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("## update")

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	if args[0] == "register" {
		return t.register(stub, args[1:])
	}

	if args[0] == "add" {
		return t.add(stub, args[1:])
	}

	if args[0] == "acquire" {
		return t.acquire(stub, args[1:])
	}

	if args[0] == "release" {
		return t.release(stub, args[1:])
	}

	// If the arguments given don’t match any function, we return an error
	return shim.Error("Unknown update action, check the second argument.")
}

func (t *ResourceManagerChaincode) register(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# register user")

	actorType, found, err := cid.GetAttributeValue(stub, model.ActorAttribute)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the type of the request owner: %v", err))
	}
	if !found {
		return shim.Error("The type of the request owner is not present")
	}

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	actorID, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
	}

	switch actorType {
	case model.ActorAdmin:
		newAdmin := model.Admin{
			Actor: model.Actor{
				ID:   actorID,
				Name: args[0],
			},
		}
		err = updateInLedger(stub, model.ObjectTypeAdmin, actorID, newAdmin)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable to register the new admin in the ledger: %v", err))
		}
		var newAdminAsByte []byte
		newAdminAsByte, err = objectToByte(newAdmin)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable convert the new admin to byte: %v", err))
		}

		fmt.Printf("Admin:\n  ID -> %s\n  Name -> %s\n", actorID, args[0])

		return shim.Success(newAdminAsByte)
	case model.ActorConsumer:
		newConsumer := model.Consumer{
			Actor: model.Actor{
				ID:   actorID,
				Name: args[0],
			},
		}
		err = updateInLedger(stub, model.ObjectTypeConsumer, actorID, newConsumer)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable to register the new consumer in the ledger: %v", err))
		}
		newConsumerAsByte, err := objectToByte(newConsumer)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable convert the new consumer to byte: %v", err))
		}

		fmt.Printf("Consumer:\n  ID -> %s\n  Name -> %s\n", actorID, args[0])

		return shim.Success(newConsumerAsByte)
	default:
		return shim.Error("The type of the request owner is unknown")
	}
}

func (t *ResourceManagerChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# add resource")

	err := cid.AssertAttributeValue(stub, model.ActorAttribute, model.ActorAdmin)
	if err != nil {
		return shim.Error(fmt.Sprintf("Only admin is allowed for the kind of request: %v", err))
	}

	if len(args) < 2 {
		return shim.Error("The number of arguments is insufficient.")
	}

	resourceID := args[0]
	if resourceID == "" {
		return shim.Error("The resource ID is empty.")
	}

	resourceDescription := args[1]
	if resourceDescription == "" {
		return shim.Error("The resource description is empty.")
	}

	resource := model.Resource{
		ID:          resourceID,
		Description: resourceDescription,
		Available:   true,
	}
	err = updateInLedger(stub, model.ObjectTypeResource, resourceID, resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to create the resource in the ledger: %v", err))
	}

	resourceAsByte, err := objectToByte(resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable convert the resource to byte: %v", err))
	}

	fmt.Printf("Resource created:\n  ID -> %s\n  Description -> %s\n", resourceID, resourceDescription)

	return shim.Success(resourceAsByte)
}

func (t *ResourceManagerChaincode) acquire(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# acquire resource")

	err := cid.AssertAttributeValue(stub, model.ActorAttribute, model.ActorConsumer)
	if err != nil {
		return shim.Error(fmt.Sprintf("Only consumer is allowed for the kind of request: %v", err))
	}

	if len(args) < 2 {
		return shim.Error("The number of arguments is insufficient.")
	}

	resourceID := args[0]
	if resourceID == "" {
		return shim.Error("The resource ID is empty.")
	}

	mission := args[1]
	if mission == "" {
		return shim.Error("The mission is empty.")
	}

	var resource model.Resource
	err = getFromLedger(stub, model.ObjectTypeResource, resourceID, &resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to find the resource in the ledger: %v", err))
	}

	if !resource.Available {
		return shim.Error(fmt.Sprintf("The resource ID '%s' is not available", resourceID))
	}

	consumerID, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
	}

	resource.Consumer = consumerID
	resource.Mission = mission
	resource.Available = false

	err = updateInLedger(stub, model.ObjectTypeResource, resourceID, resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to update the resource in the ledger: %v", err))
	}

	resourceAsByte, err := objectToByte(resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable convert the resource to byte: %v", err))
	}

	fmt.Printf("Resource acquired:\n  ID -> %s\n  Consumer ID -> %s\n  Mission -> %s\n", resourceID, consumerID, mission)

	return shim.Success(resourceAsByte)
}

func (t *ResourceManagerChaincode) release(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# release resource")

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	resourceID := args[0]
	if resourceID == "" {
		return shim.Error("The resource ID is empty.")
	}

	var resource model.Resource
	err := getFromLedger(stub, model.ObjectTypeResource, resourceID, &resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to find the resource in the ledger: %v", err))
	}

	if resource.Available {
		return shim.Error(fmt.Sprintf("The resource ID '%s' is not acquired", resourceID))
	}

	actorType, found, err := cid.GetAttributeValue(stub, model.ActorAttribute)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the type of the request owner: %v", err))
	}
	if !found {
		return shim.Error("The type of the request owner is not present")
	}

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	switch actorType {
	case model.ActorAdmin:
	case model.ActorConsumer:
		var consumerID string
		consumerID, err = cid.GetID(stub)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
		}
		if consumerID != resource.Consumer {
			return shim.Error("Unable to release a resource that you don't previously acquire")
		}
	default:
		return shim.Error("The type of the request owner is unknown")
	}

	resource.Consumer = ""
	resource.Mission = ""
	resource.Available = true

	err = updateInLedger(stub, model.ObjectTypeResource, resourceID, resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to update the resource in the ledger: %v", err))
	}

	resourceAsByte, err := objectToByte(resource)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable convert the resource to byte: %v", err))
	}

	fmt.Printf("Resource release:\n  ID -> %s\n", resourceID)

	return shim.Success(resourceAsByte)
}
