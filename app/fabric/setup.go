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

package fabric

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	caMsp "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	mspclient "github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

// Setup implementation of the Hyperledger Fabric blockchain SDK
type Setup struct {
	ChannelID           string
	ChannelConfig       string
	ChaincodeID         string
	ChaincodeGoPath     string
	ChaincodePath       string
	ChaincodeVersion    string
	OrgID               string
	OrgMspID            string
	OrgAdminUser        string
	OrdererOrgID        string
	OrdererOrgAdminUser string
	OrdererID           string
	CaID                string
	ConfigFile          string
	sdk                 *fabsdk.FabricSDK
	caClient            *caMsp.Client
}

// User stuct that allow a registered user to query and invoke the blockchain
type User struct {
	Username        string
	Fabric          *Setup
	ChannelClient   *channel.Client
	SigningIdentity msp.SigningIdentity
}

// Initialize reads the configuration stored in the Setup struct and sets up the blockchain client
func (s *Setup) Initialize() error {

	fmt.Printf("Initialise Fabric SDK...\n")

	sdk, err := fabsdk.New(config.FromFile(s.ConfigFile))
	if err != nil {
		return fmt.Errorf("failed to create new SDK: %v", err)
	}
	s.sdk = sdk

	caClient, err := caMsp.New(sdk.Context())
	if err != nil {
		return fmt.Errorf("failed to create new CA client: %v", err)
	}
	s.caClient = caClient

	fmt.Printf("Fabric SDK initialised.\n")

	return nil
}

// Install reads the configuration stored in the Setup struct and sets up the blockchain channel and chaincode
func (s *Setup) Install() error {
	fmt.Printf("Preparing contextes to create channel...\n")

	//clientContext allows creation of transactions using the supplied identity as the credential.
	clientContext := s.sdk.Context(fabsdk.WithUser(s.OrgAdminUser), fabsdk.WithOrg(s.OrdererOrgID))

	// Resource management client is responsible for managing channels (create/update channel)
	// Supply user that has privileges to create channel (in this case orderer admin)
	resMgmtClient, err := resmgmt.New(clientContext)
	if err != nil {
		return fmt.Errorf("failed to create channel management client: %s", err)
	}

	// Create channel
	err = s.createChannel(resMgmtClient)
	if err != nil {
		return fmt.Errorf("unable to create the channel: %v", err)
	}

	fmt.Printf("Preparing contextes to make peers joinning the new channel...\n")

	// Prepare context
	adminContext := s.sdk.Context(fabsdk.WithUser(s.OrgAdminUser), fabsdk.WithOrg(s.OrgID))

	// Org resource management client
	orgResMgmt, err := resmgmt.New(adminContext)
	if err != nil {
		return fmt.Errorf("failed to create new resource management client: %s", err)
	}

	// Org peers join channel
	if err = orgResMgmt.JoinChannel(s.ChannelID, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(s.OrdererID)); err != nil {
		return fmt.Errorf("org peers failed to JoinChannel: %s", err)
	}

	fmt.Printf("Peers joined the channel '%s'\n", s.ChannelID)

	err = s.createCC(orgResMgmt)
	if err != nil {
		return fmt.Errorf("unable to install and instantiate the chaincode: %v", err)
	}

	fmt.Printf("Fabric channel installed and chaincode installed/instantiated.\n")

	return nil
}

// CloseSDK allow to close communication between the initialized blockchain client and the network
func (s *Setup) CloseSDK() {
	s.sdk.Close()
}

// LogUser allow to login a user using credentials provided and retrieve the blockchain user related
func (s *Setup) LogUser(username, password string) (*User, error) {

	err := s.caClient.Enroll(username, caMsp.WithSecret(password))
	if err != nil {
		return nil, fmt.Errorf("failed to enroll identity '%s': %v", username, err)
	}

	var user User
	user.Username = username
	user.Fabric = s

	user.SigningIdentity, err = s.caClient.GetSigningIdentity(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get signing identity for '%s': %v", username, err)
	}

	clientChannelContext := s.sdk.ChannelContext(s.ChannelID, fabsdk.WithUser(username), fabsdk.WithOrg(s.OrgID), fabsdk.WithIdentity(user.SigningIdentity))

	// Channel client is used to query and execute transactions
	user.ChannelClient, err = channel.New(clientChannelContext)
	if err != nil {
		return nil, fmt.Errorf("failed to create new channel client for '%s': %v", username, err)
	}

	return &user, nil
}

// RegisterUser register a user to the Fabric CA client and into the blockchain using invoke on the chaincode
func (s *Setup) RegisterUser(username, password, userType string) error {
	fmt.Printf("Register user '%s'... \n", username)
	_, err := s.caClient.Register(&caMsp.RegistrationRequest{
		Name:           username,
		Secret:         password,
		Type:           "user",
		MaxEnrollments: -1,
		Affiliation:    "org1",
		Attributes: []caMsp.Attribute{
			{
				Name:  "actor",
				Value: userType,
				ECert: true,
			},
		},
		CAName: s.CaID,
	})
	if err != nil {
		return fmt.Errorf("unable to register user '%s': %v", username, err)
	}

	u, err := s.LogUser(username, password)
	if err != nil {
		return fmt.Errorf("unable to log user '%s' after registration: %v", username, err)
	}

	err = u.UpdateRegister()
	if err != nil {
		return fmt.Errorf("unable to add the user '%s' in the ledger: %v", username, err)
	}

	fmt.Printf("User '%s' registered.\n", username)

	return nil
}

// createChannel internal method that allow to create a channel in the blockchain network
func (s *Setup) createChannel(resMgmtClient *resmgmt.Client) error {
	fmt.Printf("Creating channel...\n")

	mspClient, err := mspclient.New(s.sdk.Context(), mspclient.WithOrg(s.OrgID))
	if err != nil {
		return err
	}
	adminIdentity, err := mspClient.GetSigningIdentity(s.OrgAdminUser)
	if err != nil {
		return err
	}

	req := resmgmt.SaveChannelRequest{
		ChannelID:         s.ChannelID,
		ChannelConfigPath: s.ChannelConfig,
		SigningIdentities: []msp.SigningIdentity{adminIdentity},
	}

	txID, err := resMgmtClient.SaveChannel(req, resmgmt.WithRetry(retry.DefaultResMgmtOpts), resmgmt.WithOrdererEndpoint(s.OrdererID))
	if err != nil {
		return err
	}

	fmt.Printf("Channel '%s' created with transaction ID '%s'\n", s.ChannelID, txID.TransactionID)
	return nil
}

// createCC internal method that allow to install and instantiate a chaincode in the blockchain network
func (s *Setup) createCC(orgResMgmt *resmgmt.Client) error {
	fmt.Printf("Install chaincode...\n")

	ccPkg, err := packager.NewCCPackage(s.ChaincodePath, s.ChaincodeGoPath)
	if err != nil {
		return err
	}

	// Install example cc to org peers
	installCCReq := resmgmt.InstallCCRequest{
		Name:    s.ChaincodeID,
		Path:    s.ChaincodePath,
		Version: s.ChaincodeVersion,
		Package: ccPkg,
	}
	_, err = orgResMgmt.InstallCC(installCCReq, resmgmt.WithRetry(retry.DefaultResMgmtOpts))
	if err != nil {
		return err
	}

	fmt.Printf("Chaincode '%s' installed (version '%s')\n", s.ChaincodeID, s.ChaincodeVersion)

	fmt.Printf("Instantiate chaincode...\n")

	// Set up chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{s.OrgMspID})

	// Org resource manager will instantiate the chaincode on channel
	resp, err := orgResMgmt.InstantiateCC(
		s.ChannelID,
		resmgmt.InstantiateCCRequest{
			Name:    s.ChaincodeID,
			Path:    s.ChaincodePath,
			Version: s.ChaincodeVersion,
			Args:    [][]byte{[]byte("init")},
			Policy:  ccPolicy,
		},
		resmgmt.WithRetry(retry.DefaultResMgmtOpts),
	)
	if err != nil {
		return err
	}

	fmt.Printf("Chaincode '%s' (version '%s') instantiated with transaction ID '%s'\n", s.ChaincodeID, s.ChaincodeVersion, resp.TransactionID)
	return nil
}
