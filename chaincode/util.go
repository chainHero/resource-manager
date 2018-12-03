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
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// objectToByte convert the given object to a slice of byte
func objectToByte(object interface{}) ([]byte, error) {
	objectAsByte, err := json.Marshal(object)
	if err != nil {
		return nil, fmt.Errorf("unable convert the object to byte: %v", err)
	}
	return objectAsByte, nil
}

// byteToObject convert the slice of byte given to the given interface object
func byteToObject(objectAsByte []byte, result interface{}) error {
	err := json.Unmarshal(objectAsByte, result)
	if err != nil {
		return fmt.Errorf("unable to convert the result to object: %v", err)
	}
	return nil
}

// getFromLedger retrieve an object from the ledger
func getFromLedger(stub shim.ChaincodeStubInterface, objectType string, id string, result interface{}) error {
	key, err := stub.CreateCompositeKey(objectType, []string{id})
	if err != nil {
		return fmt.Errorf("unable to create the object key for the ledger: %v", err)
	}
	resultAsByte, err := stub.GetState(key)
	if err != nil {
		return fmt.Errorf("unable to retrieve the object in the ledger: %v", err)
	}
	if resultAsByte == nil {
		return fmt.Errorf("the object doesn't exist in the ledger")
	}
	err = byteToObject(resultAsByte, result)
	if err != nil {
		return fmt.Errorf("unable to convert the result to object: %v", err)
	}
	return nil
}

// updateInLedger update an object in the ledger
func updateInLedger(stub shim.ChaincodeStubInterface, objectType string, id string, object interface{}) error {
	key, err := stub.CreateCompositeKey(objectType, []string{id})
	if err != nil {
		return fmt.Errorf("unable to create the object key for the ledger: %v", err)
	}

	objectAsByte, err := objectToByte(object)
	if err != nil {
		return err
	}
	err = stub.PutState(key, objectAsByte)
	if err != nil {
		return fmt.Errorf("unable to put the object in the ledger: %v", err)
	}
	return nil
}

// deleteFromLedger delete an object in the ledger
func deleteFromLedger(stub shim.ChaincodeStubInterface, objectType string, id string) error {
	key, err := stub.CreateCompositeKey(objectType, []string{id})
	if err != nil {
		return fmt.Errorf("unable to create the object key for the ledger: %v", err)
	}
	err = stub.DelState(key)
	if err != nil {
		return fmt.Errorf("unable to delete the object in the ledger: %v", err)
	}
	return nil
}
