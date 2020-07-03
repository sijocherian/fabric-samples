/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)


// read asset information from collection
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {

	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, assetID) //get the asset from chaincode state
	if err != nil {
			return nil, fmt.Errorf("failed to read from asset %s", err.Error())
		}
		if assetJSON == nil {
			return nil, fmt.Errorf("%s does not exist", assetID)
		}

		asset := new(Asset)
	err = json.Unmarshal(assetJSON, asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}


	return asset, nil

}

// read asset private details in organization specific collection
func (s *SmartContract) ReadAssetPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*AssetPrivateDetails, error) {

	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) //get the asset from chaincode state
		if err != nil {
			return nil, fmt.Errorf("failed to read from asset details %s", err.Error())
		}
		if assetDetailsJSON == nil {
			return nil, fmt.Errorf("%s does not exist", assetID)
		}

		assetDetails := new(AssetPrivateDetails)
	err = json.Unmarshal(assetDetailsJSON, assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	return assetDetails, nil
}

// get identity from asset agreement from collection
func (s *SmartContract) ReadTransferAgreement(ctx contractapi.TransactionContextInterface, assetID string) (string, error) {

	// create composite key
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey("transferAgreement",[]string{assetID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %s", err.Error())
	}

	buyerIdentity, err := ctx.GetStub().GetPrivateData(assetCollection, transferAgreeKey) //get the identity from collection
	if err != nil {
			return "", fmt.Errorf("failed to read from asset %s", err.Error())
		}
		if buyerIdentity == nil {
			return "", fmt.Errorf("%s does not exist", assetID)
		}

	return string(buyerIdentity), nil

}

// ===========================================================================================
// GetAssetByRange performs a range query based on the start and end keys provided. Range
// queries can be used to read data from private data collections, but can not be used in
// a transaction that also writes to private data.

// ===========================================================================================
func (s *SmartContract) GetAssetByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange("assetCollection", startKey, endKey)
	if err != nil {
			return nil, err
	}
	defer resultsIterator.Close()

	results := []Asset{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		newAsset := new(Asset)

		err = json.Unmarshal(response.Value, newAsset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
		}

		results = append(results, *newAsset)
	}

	return results, nil

}


// =======Rich queries =========================================================================
// Two examples of rich queries are provided below (parameterized query and ad hoc query).
// Rich queries pass a query string to the state database.
// Rich queries are only supported by state database implementations
//  that support rich query (e.g. CouchDB).
// The query string is in the syntax of the underlying state database.
// With rich queries there is no guarantee that the result set hasn't changed between
//  endorsement time and commit time, aka 'phantom reads'.
// Therefore, rich queries should not be used in update transactions, unless the
// application handles the possibility of result set changes between endorsement and commit time.
// Rich queries can be used for point-in-time queries against a peer.
// ============================================================================================

// ===== Example: Parameterized rich query =================================================
// QueryAssetByOwner queries for assets based on a passed in owner.
// This is an example of a parameterized query where the query logic is baked into the chaincode,
// and accepting a single query parameter (owner).
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (s *SmartContract) QueryAssetByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]Asset, error) {

	ownerString  := strings.ToLower(owner)

	queryString := fmt.Sprintf("{\"selector\":{\"docType\":\"asset\",\"owner\":\"%s\"}}", ownerString)

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)
	if err != nil {
			return nil, err
	}
	return queryResults, nil
}

// ===== Example: Ad hoc rich query ========================================================
// QueryAssets uses a query string to perform a query for assets.
// Query string matching state database syntax is passed in and executed as is.
// Supports ad hoc queries that can be defined at runtime by the client.
// If this is not desired, follow the QueryAssetByOwner example for parameterized queries.
// Only available on state databases that support rich query (e.g. CouchDB)
// =========================================================================================
func (s *SmartContract) QueryAssets(ctx contractapi.TransactionContextInterface, queryString string) ([]Asset, error) {

	queryResults, err := s.getQueryResultForQueryString(ctx, queryString)
	if err != nil {
			return nil, err
	}
	return queryResults, nil
}

// getQueryResultForQueryString executes the passed in query string.
func (s *SmartContract) getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]Asset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult("assetCollection", queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []Asset{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		newAsset := new(Asset)

		err = json.Unmarshal(response.Value, newAsset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
		}

		results = append(results, *newAsset)
	}
	return results, nil
}
