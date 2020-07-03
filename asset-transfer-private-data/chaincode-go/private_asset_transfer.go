/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// Asset describes main asset details that are visible to all organizations
type Asset struct {
	Type 		 			 string `json:"type"`  //Type is used to distinguish the various types of objects in state database
	ID             string `json:"asset_id"`
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
}

// AssetPrivateDetails describes details that are private to owners
type AssetPrivateDetails struct {
	ID             string `json:"asset_id"`
	AppraisedValue int    `json:"appraised_value"`
}

const assetCollection = "assetCollection"

type SmartContract struct {
	contractapi.Contract
}

// CreateAsset creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface) error {

	// Get new asset from transient map
	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: " + err.Error())
	}

	// Asset properties are private, therefore they get passed in transient field
	transientAssetJSON, ok := transMap["asset_properties"]
	if !ok {
		return fmt.Errorf("asset not found in the transient map")
	}

	type assetTransientInput struct {
		Type 		 				string `json:"type"`  //Type is used to distinguish the various types of objects in state database
		ID             	string `json:"asset_id"`
		Color          	string `json:"color"`
		Size           	int    `json:"size"`
		AppraisedValue 	int    `json:"appraised_value"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	if len(assetInput.Type) == 0 {
		return fmt.Errorf("type field must be a non-empty string")
	}
	if len(assetInput.ID) == 0 {
		return fmt.Errorf("asset_id field must be a non-empty string")
	}
	if len(assetInput.Color) == 0 {
		return fmt.Errorf("color field must be a non-empty string")
	}
	if assetInput.Size <= 0 {
		return fmt.Errorf("size field must be a positive integer")
	}
	if assetInput.AppraisedValue <= 0 {
		return fmt.Errorf("AppraisedValue field must be a positive integer")
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.ID)
	if err != nil {
		return fmt.Errorf("Failed to get asset: " + err.Error())
	} else if assetAsBytes != nil {
		fmt.Println("This asset already exists: " + assetInput.ID)
		return fmt.Errorf("This asset already exists: " + assetInput.ID)
	}

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	}

	// Make submitting client the owner
	asset := &Asset{
		Type: 			assetInput.Type,
		ID:       	assetInput.ID,
		Color:      assetInput.Color,
		Size:       assetInput.Size,
		Owner:      clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	// Save asset to private data collection
	err = ctx.GetStub().PutPrivateData(assetCollection, assetInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put Asset: %s", err.Error())
	}

	// Save asset details to collection visible to owning organization
	assetPrivateDetails := &AssetPrivateDetails{
		ID:       			 	assetInput.ID,
		AppraisedValue:   assetInput.AppraisedValue,
	}

	assetPrivateDetailsAsBytes, err := json.Marshal(assetPrivateDetails) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	// get collection name for this organization
	orgCollection, err := getCollectionName(ctx)

	// put asset appraised value into owners org specific private data collection
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %s", err.Error())
	}
	return nil
}

// AgreeToTransfer is used by the potential buyer of the asset to agree to the
// asset value. The agreed to appraisal value is stored in the buying orgs
// org specifc collection, while the the buyer client ID is stored in the asset collection
// using a composite key
func (s *SmartContract) AgreeToTransfer(ctx contractapi.TransactionContextInterface) error {

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	}

	// value is private, therefore it gets passed in transient field
	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: " + err.Error())
	}

	// persist the JSON bytes as-is so that there is no risk of nondeterministic marshaling.
	valueJSONasBytles, ok := transMap["asset_value"]
	if !ok {
		return fmt.Errorf("asset_value key not found in the transient map")
	}

	// unmarshal the tranisent map to get the asset ID.
	var valueJSON AssetPrivateDetails
	err = json.Unmarshal(valueJSONasBytles, &valueJSON)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	// do some error checking since we get the chance
	if len(valueJSON.ID) == 0 {
		return fmt.Errorf("asset_id field must be a non-empty string")
	}
	if valueJSON.AppraisedValue <= 0 {
		return fmt.Errorf("AppraisedValue field must be a positive integer")
	}

	// get collection name for this organization
	orgCollection, err := getCollectionName(ctx)

	// put agreed value in the org specifc private data collection
	err = ctx.GetStub().PutPrivateData(orgCollection, valueJSON.ID, valueJSONasBytles)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %s", err.Error())
	}

	// create agreeement where you indicate the identity has agreed to purchase
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey("transferAgreement",[]string{valueJSON.ID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %s", err.Error())
	}

	err = ctx.GetStub().PutPrivateData(assetCollection, transferAgreeKey, []byte(clientID))
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %s", err.Error())
	}

	return nil
}


// transfer a asset by setting a new owner ID on the asset
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface) error {

	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: " + err.Error())
	}

	// Asset properties are private, therefore they get passed in transient field
	transientTransferJSON, ok := transMap["asset_owner"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	type assetTransferTransientInput struct {
		ID  			string 	`json:"asset_id"`
		BuyerMSP 	string 	`json:"buyer_msp"`
	}

	var assetTransferInput assetTransferTransientInput
	err = json.Unmarshal(transientTransferJSON, &assetTransferInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	if len(assetTransferInput.ID) == 0 {
		return fmt.Errorf("asset_id field must be a non-empty string")
	}
	if len(assetTransferInput.BuyerMSP) == 0 {
		return fmt.Errorf("owner field must be a non-empty string")
	}

	// Read asset from the private data collection
	asset, err := s.ReadAsset(ctx, assetTransferInput.ID)
		if err != nil {
			return fmt.Errorf("failed to get asset: %s", err.Error())
		}

	// verify transfer details and transfer owner
	err = s.verifyAgreement(ctx, assetTransferInput.ID, asset.Owner, assetTransferInput.BuyerMSP)
		if err != nil {
			return fmt.Errorf("failed transfer verification: %s", err.Error())
		}

	buyerID, err := s.ReadTransferAgreement(ctx, assetTransferInput.ID)

	// Transfer asset in private data collection to new owner
	asset.Owner = buyerID

	assetJSONasBytes, _ := json.Marshal(asset)
	err = ctx.GetStub().PutPrivateData(assetCollection, assetTransferInput.ID, assetJSONasBytes) //rewrite the asset
		if err != nil {
				return err
		}

	// get collection name for this organization
	ownersCollection, err := getCollectionName(ctx)

	// delete the marble appraised value from this organiztion's private data collection
	err = ctx.GetStub().DelPrivateData(ownersCollection, assetTransferInput.ID)
		if err != nil {
				return err
		}

	// delete the transfer agreement from the asset collection
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey("transferAgreement",[]string{assetTransferInput.ID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %s", err.Error())
	}

	err = ctx.GetStub().DelPrivateData(assetCollection,transferAgreeKey)
		if err != nil {
				return err
		}

	return nil

}

// verifyAgreement is an internal helper function used by TransferAsset to verify
// that the transfer is being initiated by the owner and that the buyer has agreed
// to the same appraisal value as the owner
func (s *SmartContract) verifyAgreement(ctx contractapi.TransactionContextInterface, assetID string, owner string, buyerMSP string) error {

	// Check 1: verify that the transfer is being initiatied by the owner

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	}

	if clientID != owner {
			return fmt.Errorf("error: submitting client identity does not own asset")
		}

	// Check 2: verify that the buyer has agreed to the appraised value

	// get collection names

	collectionOwner, err := getCollectionName(ctx) // get buyers collection

	collectionBuyer := buyerMSP + "PrivateCollection" // get buyers collection

	// get hash of owners agreed to value
	ownerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionOwner, assetID)
		if err != nil {
			return fmt.Errorf("failed to hash of appraised value from owners collection %s: %s", collectionOwner ,err.Error())
		}
		if ownerAppraisedValueHash == nil {
			return fmt.Errorf("hash of appraised value for %s does not exist in collection %s", assetID, collectionOwner)
		}

		// get hash of buyers agreed to value
	buyerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionBuyer, assetID)
		if err != nil {
			return fmt.Errorf("failed to hash of appraised value from buyer collection %s: %s", collectionBuyer ,err.Error())
		}
		if buyerAppraisedValueHash == nil {
			return fmt.Errorf("hash of appraised value for %s does not exist in collection %s", assetID, collectionBuyer)
		}

	// verify that the two hashes match
	if !bytes.Equal(ownerAppraisedValueHash, buyerAppraisedValueHash) {
		return fmt.Errorf("hash for appraised value for owner %x does not value for seller %x", ownerAppraisedValueHash, buyerAppraisedValueHash)
	}

	return nil
}

// DeleteAsset can be used by the owner of the asset to delete the asset
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface) error {

	transMap, err := ctx.GetStub().GetTransient()
		if err != nil {
			return fmt.Errorf("Error getting transient: " + err.Error())
		}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transMap["asset_delete"]
		if !ok {
			return fmt.Errorf("asset to delete not found in the transient map")
		}

	type assetDelete struct {
			ID string `json:"asset_id"`
		}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
		}

	if len(assetDeleteInput.ID) == 0 {
		return fmt.Errorf("ID field must be a non-empty string")
	}

	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetDeleteInput.ID) //get the asset from chaincode state
		if err != nil {
			return fmt.Errorf("failed to read asset: %s", err.Error())
		}
		if valAsbytes == nil {
			return fmt.Errorf("asset private details does not exist: %s",  assetDeleteInput.ID)
		}

	var assetToDelete Asset
	err = json.Unmarshal([]byte(valAsbytes), &assetToDelete)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	// delete the asset from state
	err = ctx.GetStub().DelPrivateData(assetCollection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("Failed to delete state:" + err.Error())
	}

	// Finally, delete private details of asset

	ownerCollection, err := getCollectionName(ctx) // get owners collection

	err = ctx.GetStub().DelPrivateData(ownerCollection, assetDeleteInput.ID) // delete the asset
	if err != nil {
			return err
	}

	return nil

}

// DeleteProposal can be used by a user of a rejected transfer properties to
// remove the propsal from the asset collection and the org specific collection
func (s *SmartContract) DeleteProposal(ctx contractapi.TransactionContextInterface) error {

	transMap, err := ctx.GetStub().GetTransient()
		if err != nil {
			return fmt.Errorf("Error getting transient: " + err.Error())
		}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transMap["agree_delete"]
		if !ok {
			return fmt.Errorf("asset to delete not found in the transient map")
		}

	type assetDelete struct {
			ID string `json:"asset_id"`
		}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
		}

	if len(assetDeleteInput.ID) == 0 {
		return fmt.Errorf("ID field must be a non-empty string")
	}

	// delete private details of agreement

	orgCollection, err := getCollectionName(ctx) // get owners collection

	err = ctx.GetStub().DelPrivateData(orgCollection, assetDeleteInput.ID) // delete the asset
	if err != nil {
			return err
	}

	// delete transfer agreement record

	tranferAgreeKey, err := ctx.GetStub().CreateCompositeKey("transferAgreement",[]string{assetDeleteInput.ID}) // create composite key
		if err != nil {
			return fmt.Errorf("failed to create composite key: %s", err.Error())
		}

	err = ctx.GetStub().DelState(tranferAgreeKey) // remove agreement from state
		if err != nil {
			return err
		}

	return nil

}

// Return the ID of submitting client identity. This will be used to identify
// and verify the owner on chain
func (s *SmartContract) ReturnID(ctx contractapi.TransactionContextInterface) (string, error) {

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	}
	return clientID, nil
}

// getCollectionName is an internal helper function to get collection of submitting client identity
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified OrgID: %s", err.Error())
	}

	// verify that the client is submitting request to peer in their organization
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting peer's orgID: %s", err.Error())
	}

	if clientMSPID != peerOrgID {
		return "", fmt.Errorf("client from org %s is not authorized to read or write private data on org %s peer", clientMSPID, peerOrgID)
	}

	// create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error creating private mables chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting private mables chaincode: %s", err.Error())
	}
}
