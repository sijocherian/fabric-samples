# Private data transfer scenario

The private data transfer smart contract demonstrates a simple asset transfer that uses private data collections. All data is stored in private data collections and cannot be read by all members of the channel. The chaincode uses the `collections_config.json` file to deploy three private data collections:

- The `assetCollection` is stored on the peers of Org1 and Org2 and is used to store the properties of an asset that can be accessed by both Org1 and Org2. This collection is used to store the main details of the asset, such as the size, color, and the owner.
- `Org1MSPDetailsCollection` is stored only on the peers of Org1, while the `Org2MSPDetailsCollection` is only stored on the peers of Org2. These collections are used to store the appraised value of the asset. These collections are used to store the appraisal value of the asset.
- The asset is owned by the client application that creates the asset. The chaincode uses the `GetID()` function to fetch the information of the ID that submitted the request, and assigns that client application as the owner of the asset. The appraised value is stored in the Details collection of the organization that owns the asset. For example, if a user from Org1 uses the smart contract to create the asset, the appraisal value is stored in the `Org1MSPDetailsCollection`.
- If the other organization wants to purchase the asset, the can use the smart contract to create a transfer agreement. The buyer needs to agree to a value for the asset. The smart contract will store this value in the collection of the organization that agrees to buy the asset. The transfer agreement will also store the client ID of the buyer in the `assetCollection`.
- After another user has agreed to buy the asset, the owner can transfer the asset. The transfer function will check that user transferring the asset is the asset owner. The user will also use the `GetPrivateDataHash()` function to check that the purchaser of the asset has agreed to the same appraisal value as the owner. If the buyer has agreed to the same price, the transfer function will get the client ID of the buyer from the transfer agreement, and updates the owner of the asset in the `assetCollection`.

This smart contract is meant to introduce users to how to use private data collections. For an example of a more realistic asset transfer scenario, see the [secure asset transfer smart contract](link).

## Deploy the smart contract to the test network

You can use the Fabric test network to run the private data transfer scenario. Open a command terminal and navigate to test network directory in your local clone of the `fabric-samples`. We will operate from the `test-network` directory for the remainder of the tutorial.
```
cd fabric-samples/test-network
```

The test network contains two peer organizations. We will deploy the test network using certificate authorities, so we can use a CA for each organization. The script will also a single channel named `mychannel` with Org1 and Org2 as channel members.

```
./network.sh up createChannel -ca
```

## Deploy the smart contract to the channel

You can use the following steps to deploy the smart contract to the channel.

### Install and approve the chaincode as Org1

Set the following environment variables to operate the `peer` CLI as the Org1 admin:
```
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:7051
```

Run the following command to package the private asset transfer chaincode:
```
peer lifecycle chaincode package private_transfer.tar.gz --path ../asset-transfer-private-data/chaincode-go --lang golang --label private_transfer_1
```

The command creates a chaincode package named `private_transfer.tar.gz`. We can now install this package on the Org1 peer:
```
peer lifecycle chaincode install private_transfer.tar.gz
```

You will need the chaincode package ID in order to approve the chaincode definition. You can find the package ID by querying your peer:
```
peer lifecycle chaincode queryinstalled
```
Save the package ID as an environment variable. The package ID will not be the same for all users, so need to use the result that was returned by the previous command:
```
export PACKAGE_ID=private_transfer_1:5bfc5d3a7ca7110d7f69473eadf09f8f8dde0a24d1d3fceb4f7c05bf355bb990
```
You can now approve the chaincode as Org1. This command includes a path to the collection definition file.
```
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name private_transfer --version 1 --package-id $PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --collections-config ../asset-transfer-private-data/chaincode-go/collections_config.json --signature-policy "OR('Org1MSP.peer','Org2MSP.peer')"
```

Note we are approving a chaincode endorsement policy of `"OR('Org1MSP.peer','Org2MSP.peer')"`. This allows either organization to create a asset without receiving an endorsement from the other organization.


### Install and approve the chaincode as Org2

We can now install and approve the chaincode as Org2. Set the following environment variables to operate as the Org2 admin:
```
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:9051
```

Because the chaincode is already packaged on our local machine, we can go ahead and install the chaincode on the Org2 peer:`
```
peer lifecycle chaincode install private_transfer.tar.gz
```

We can now approve the chaincode as the Org2 admin:
```
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name private_transfer --version 1 --package-id $PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --collections-config ../asset-transfer-private-data/chaincode-go/collections_config.json --signature-policy "OR('Org1MSP.peer','Org2MSP.peer')"
```

### Commit the chaincode definition the channel

Now that a majority (2 out of 2) of channel members have approved the chaincode definition, Org2 can commit the chaincode definition to deploy the chaincode to the channel:
```
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name private_transfer --version 1 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --collections-config ../asset-transfer-private-data/chaincode-go/collections_config.json --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt --signature-policy "OR('Org1MSP.peer','Org2MSP.peer')"
```
We are now ready use the private asset transfer smart contract.

## Register identities

The private data transfer smart contract supports ownership by individual identities. In our scenario, the owner of the asset will belong to Org1, while the buyer will belong to Org2. To highlight this, we will register two new identities with both organizations.

First, we will use the Org1 CA to create the owner identity. Set the Fabric CA client home the Org1 CA admin (this identity was generated by the script):
```
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org1.example.com/
```

You can register a new owner client identity using the `fabric-ca-client` tool:
```
fabric-ca-client register --caname ca-org1 --id.name owner --id.secret ownerpw --id.type client --tls.certfiles ${PWD}/organizations/fabric-ca/org1/tls-cert.pem
```

We can now enroll using the enroll name and secret to generate the identity crypto material.
```
fabric-ca-client enroll -u https://owner:ownerpw@localhost:7054 --caname ca-org1 -M ${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp --tls.certfiles ${PWD}/organizations/fabric-ca/org1/tls-cert.pem
```

Run the command below to copy the Node OU configuration file into the owner identity MSP folder.
```
cp ${PWD}/organizations/peerOrganizations/org1.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp/config.yaml
```

We can now use the Org2 CA to create the buyer identity. Set the Fabric CA client home the Org2 CA admin:
```
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org2.example.com/
```

You can register a new owner client identity using the `fabric-ca-client` tool:
```
fabric-ca-client register --caname ca-org2 --id.name buyer --id.secret buyerpw --id.type client --tls.certfiles ${PWD}/organizations/fabric-ca/org2/tls-cert.pem
```

We can now enroll using the enroll name and secret to generate the identity crypto material.
```
fabric-ca-client enroll -u https://buyer:buyerpw@localhost:8054 --caname ca-org2 -M ${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp --tls.certfiles ${PWD}/organizations/fabric-ca/org2/tls-cert.pem
```

Run the command below to copy the Node OU configuration file into the buyer identity MSP folder.
```
cp ${PWD}/organizations/peerOrganizations/org2.example.com/msp/config.yaml ${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp/config.yaml
```

## Create an asset

We can now use the smart contract to create an asset that is owned by the owner identity from Org1. Use the following environment variables to operate the `peer` CLI as the owner identity from Org1.

```
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:7051
```

Run the following command to define the asset properties:3
```
export asset_PROPERTIES=$(echo -n "{\"object_type\":\"asset\",\"asset_id\":\"asset1\",\"color\":\"green\",\"size\":20,\"appraisedValue\":100}" | base64 | tr -d \\n)
```

We can now invoke the chaincode to create an asset that belongs to the owner identity:
```
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"CreateAsset","Args":[]}' --transient "{\"asset_properties\":\"$asset_PROPERTIES\"}"
```

We can can read the asset properties by querying the `assetCollection` private data collection:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAsset","Args":["asset1"]}'
```

When successful, the command will return the following result:
```
"{\"object_type\":\"asset\",\"asset_id\":\"asset1\",\"color\":\"green\",\"size\":20,\"owner\":\"eDUwOTo6Q049b3duZXIsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw==\"}"
```

The `"owner"` of the asset is the identity that invoked the chaincode to create the asset, as identified by the common name and issuer of the identities certificate, which is then base64 encoded. You can see that information by base64 decoding the owner string:
```
echo eDUwOTo6Q049b3duZXIsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMS5leGFtcGxlLmNvbSxPPW9yZzEuZXhhbXBsZS5jb20sTD1EdXJoYW0sU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUw | base64 --decode
```

The result will show the common name and issuer of the owner certificate:
```
x509::CN=owner,OU=client,O=Hyperledger,ST=North Carolina,C=US::CN=ca.org1.example.com,O=org1.example.com,L=Durham,ST=North Carolina,C=Umacbook-air:test-network
```

The owner can also read the marble details that are stored in the `Org1MSPDetailsCollection` stored on the Org1 peer:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAssetPrivateDetails","Args":["Org1MSPDetailsCollection","asset1"]}'
```
The query will return the price of the asset:
```
"{\"ID\":\"asset1\",\"appraisedValue\":100}"
```

### Buyer from Org2 agrees to buy the asset

The buyer that belongs to Org2 is interested in buying the asset. Set the following environment variables to operate as the buyer of the asset:

```
export CORE_PEER_LOCALMSPID="Org2MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/buyer@org2.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:9051
```

Now that we are operating as a member of Org2, we should not that the asset details are not currently stored in Org2 private data collection:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAssetPrivateDetails","Args":["Org2MSPDetailsCollection","asset1"]}'
```
The buyer only finds that asset1 does exist in his collection:
```
Error: endorsement failure during invoke. response: status:500 message:"asset1 does not exist"
```

Nor is a member of Org2 able to read the Org1 private data collection:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAssetPrivateDetails","Args":["Org1MSPDetailsCollection","asset1"]}'
```
By setting `"memberOnlyRead": true` in the collection configuration file, only a member of Org1 can read the collection. A member from Org2 only gets the following response.
```
Error: endorsement failure during query. response: status:500 message:"failed to read from asset details GET_STATE failed: transaction ID: f695f08e71667d7124d7779d2312b21b67320b37f528d453bbd117ffd87ec86b: tx creator does not have read access permission on privatedata in chaincodeName:private_transfer collectionName: Org1MSPDetailsCollection"
```

To purchase the asset, the buyer needs to agree to the price set by the owner. The buyer can then store that price in the `Org2MSPDetailsCollection` private details collection. Run the following command to agree to the appraised value of 100:
```
export asset_PRICE=$(echo -n "{\"asset_id\":\"asset1\",\"appraisedValue\":100}" | base64 | tr -d \\n)
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"AgreeToPrice","Args":["asset1"]}' --transient "{\"asset_price\":\"$asset_PRICE\"}"
```

The buyer can now query the price that they agreed in the Org2 private data collection:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAssetPrivateDetails","Args":["Org2MSPDetailsCollection","asset1"]}'
```
The invoke will return the following value:
```
{"asset_id":"asset1","appraisedValue":100}
```

To purchase the asset, the buyer needs to pass their identity out of band. The buyer can use the return ID function for that purpose:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReturnID","Args":[]}'
```

The query returns the following string:
```
eDUwOTo6Q049YnV5ZXIsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL
```

## Transfer the asset to Org2

Now that buyer has agreed to buy the asset for appraised value, Org1 can transfer the asset to Org2. Set the following environment variables to operate as Org1:
```
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/owner@org1.example.com/msp
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_ADDRESS=localhost:7051
```


To transfer the asset, the owner needs to pass the MPS ID and identity of the new owner to the transfer function using transient data.
```
export asset_Owner=$(echo -n "{\"asset_id\":\"asset1\",\"buyer_id\":\"eDUwOTo6Q049YnV5ZXIsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL\",\"buyer_msp\":\"Org2MSP\"}" | base64 | tr -d \\n)
```

Operate from the Org1 terminal. The owner of the asset needs to initiate the transfer. Note that the command below uses the `--peerAddresses` flag to target the peers of both Org1 and Org2. Both organizations need to endorse the transfer.

```
peer chaincode invoke -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"TransferAsset","Args":[]}' --transient "{\"asset_owner\":\"$asset_Owner\"}" --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt
```
You can query `asset1` to see the results of the transfer.
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAsset","Args":["asset1"]}'
```

The results will show that the buyer identity now owns the asset:

```
{"object_type":"asset","asset_id":"asset1","color":"green","size":20,"owner":"eDUwOTo6Q049YnV5ZXIsT1U9Y2xpZW50LE89SHlwZXJsZWRnZXIsU1Q9Tm9ydGggQ2Fyb2xpbmEsQz1VUzo6Q049Y2Eub3JnMi5leGFtcGxlLmNvbSxPPW9yZzIuZXhhbXBsZS5jb20sTD1IdXJzbGV5LFNUPUhhbXBzaGlyZSxDPVVL"}
```

You can also confrirm that transfer removed the private details from the Org1 collection:
```
peer chaincode query -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem -C mychannel -n private_transfer -c '{"function":"ReadAssetPrivateDetails","Args":["Org1MSPDetailsCollection","asset1"]}'
```
You query will return the following result:
```
Error: endorsement failure during query. response: status:500 message:"asset1 does not exist"
```

## Clean up

When you are finished, you can bring down the test network. The command will remove all the nodes of the test network, and delete any ledger data that you created:

```
./network.sh down
```
