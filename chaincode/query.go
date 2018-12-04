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
	"time"
)

// query function that handle every readonly in the ledger
func (t *ResourceManagerChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("## query")

	// Check whether the number of arguments is sufficient
	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	if args[0] == "admin" {
		return t.admin(stub, args[1:])
	}

	if args[0] == "consumer" {
		return t.consumer(stub, args[1:])
	}

	if args[0] == "resources" {
		return t.resources(stub, args[1:])
	}

	if args[0] == "resource" {
		return t.resource(stub, args[1:])
	}

	// If the arguments given don’t match any function, we return an error
	return shim.Error("Unknown query action, check the second argument.")
}

func (t *ResourceManagerChaincode) admin(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# admin information")

	err := cid.AssertAttributeValue(stub, model.ActorAttribute, model.ActorAdmin)
	if err != nil {
		return shim.Error(fmt.Sprintf("Only admin is allowed for the kind of request: %v", err))
	}

	adminID, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
	}
	var admin model.Admin
	err = getFromLedger(stub, model.ObjectTypeAdmin, adminID, &admin)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to retrieve admin in the ledger: %v", err))
	}
	adminAsByte, err := objectToByte(admin)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable convert the admin to byte: %v", err))
	}

	return shim.Success(adminAsByte)
}

func (t *ResourceManagerChaincode) consumer(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# consumer information")

	err := cid.AssertAttributeValue(stub, model.ActorAttribute, model.ActorConsumer)
	if err != nil {
		return shim.Error(fmt.Sprintf("Only consumer is allowed for the kind of request: %v", err))
	}

	consumerID, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
	}
	var consumer model.Consumer
	err = getFromLedger(stub, model.ObjectTypeConsumer, consumerID, &consumer)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to retrieve consumer in the ledger: %v", err))
	}
	clientAsByte, err := objectToByte(consumer)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable convert the consumer to byte: %v", err))
	}

	return shim.Success(clientAsByte)
}

func (t *ResourceManagerChaincode) resources(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# resources list")

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	iterator, err := stub.GetStateByPartialCompositeKey(model.ObjectTypeResource, []string{})
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to retrieve the list of resource in the ledger: %v", err))
	}

	actorType, found, err := cid.GetAttributeValue(stub, model.ActorAttribute)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the type of the request owner: %v", err))
	}
	if !found {
		return shim.Error("The type of the request owner is not present")
	}

	actorID, err := cid.GetID(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to identify the ID of the request owner: %v", err))
	}

	filter := args[0]
	resources := make([]model.Resource, 0)

	for iterator.HasNext() {
		keyValueState, errIt := iterator.Next()
		if errIt != nil {
			return shim.Error(fmt.Sprintf("Unable to retrieve a resource in the ledger: %v", errIt))
		}
		var resource model.Resource
		err = byteToObject(keyValueState.Value, &resource)
		if err != nil {
			return shim.Error(fmt.Sprintf("Unable to convert a resource: %v", err))
		}
		if isResourceCanBeReturned(actorID, actorType, filter, &resource) {
			resources = append(resources, resource)
		}
	}

	resourcesAsByte, err := objectToByte(resources)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to convert the resource list to byte: %v", err))
	}

	return shim.Success(resourcesAsByte)
}

// isResourceCanBeReturned check if the resource can be return to the given actor and filter given.
func isResourceCanBeReturned(actorID string, actorType string, filter string, resource *model.Resource) bool {
	// If the request owner is a consumer, we give only available resources or its  previously acquired
	if model.ActorConsumer == actorType && !resource.Available && resource.Consumer != actorID {
		return false
	}
	if filter == model.ResourcesFilterOnlyAvailable && !resource.Available {
		return false
	}
	if filter == model.ResourcesFilterOnlyUnavailable && resource.Available {
		return false
	}
	return true
}

func (t *ResourceManagerChaincode) resource(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	fmt.Println("# resource detail")

	err := cid.AssertAttributeValue(stub, model.ActorAttribute, model.ActorAdmin)
	if err != nil {
		return shim.Error(fmt.Sprintf("Only admin is allowed for the kind of request: %v", err))
	}

	if len(args) < 1 {
		return shim.Error("The number of arguments is insufficient.")
	}

	resourceID := args[0]
	if resourceID == "" {
		return shim.Error("The resource ID is empty.")
	}

	key, err := stub.CreateCompositeKey(model.ObjectTypeResource, []string{resourceID})
	if err != nil {
		return shim.Error(fmt.Sprintf("unable to create the object key for the ledger: %v", err))
	}
	iterator, err := stub.GetHistoryForKey(key)
	if err != nil {
		return shim.Error(fmt.Sprintf("unable to retrieve an history resource in the ledger: %v", err))
	}

	var resourceHistories model.ResourceHistories
	for iterator.HasNext() {
		historyState, errIt := iterator.Next()
		if errIt != nil {
			return shim.Error(fmt.Sprintf("unable to retrieve a resource history in the ledger: %v", errIt))
		}
		var resourceHistory model.ResourceHistory
		resourceHistory.Deleted = historyState.GetIsDelete()
		if !resourceHistory.Deleted {
			err = byteToObject(historyState.GetValue(), &resourceHistory.Resource)
			if err != nil {
				return shim.Error(fmt.Sprintf("unable to convert the resource history value to a valid resource: %v", err))
			}
		}
		resourceHistory.Transaction = historyState.GetTxId()
		timestamp := historyState.GetTimestamp()
		resourceHistory.Time = time.Unix(timestamp.Seconds, int64(timestamp.Nanos))
		resourceHistories = append(resourceHistories, resourceHistory)
	}

	if len(resourceHistories) <= 0 {
		return shim.Error(fmt.Sprintf("Unable to found an history for the resource ID given: %v", err))
	}

	resourcesHistoryAsByte, err := objectToByte(resourceHistories)
	if err != nil {
		return shim.Error(fmt.Sprintf("Unable to convert the resource histories to byte: %v", err))
	}

	return shim.Success(resourcesHistoryAsByte)
}
