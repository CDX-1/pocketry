package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/bytemare/opaque"
	bytemare "github.com/bytemare/opaque"
)

const DefaultOpaqueServerIdentity = "pocketry-server"

type BytemareOpaqueServer struct {
	conf     *bytemare.Configuration
	serverID []byte
	skm      *bytemare.ServerKeyMaterial
}

type PendingOpaqueRegistrationState struct {
	CredentialIdentifier string `json:"credential_identifier"`
	ClientIdentity       string `json:"client_identity"`
}

type PendingOpaqueLoginState struct {
	ExpectedClientMAC string `json:"expected_client_mac"`
}

type StoredOpaqueRecord struct {
	CredentialIdentifier string `json:"credential_identifier"`
	ClientIdentity       string `json:"client_identity"`
	RegistrationRecord   string `json:"registration_record"`
}

func NewBytemareOpaqueServer(
	serverIdentity string,
	serializedServerKeyMaterial []byte,
) (*BytemareOpaqueServer, error) {
	conf := bytemare.DefaultConfiguration()

	if serverIdentity == "" {
		serverIdentity = DefaultOpaqueServerIdentity
	}

	if len(serializedServerKeyMaterial) == 0 {
		return nil, fmt.Errorf("OPAQUE server key material is required")
	}

	skm, err := conf.DecodeServerKeyMaterial(serializedServerKeyMaterial)
	if err != nil {
		return nil, fmt.Errorf("decode OPAQUE server key material: %w", err)
	}

	skm.Identity = []byte(serverIdentity)

	server, err := conf.Server()
	if err != nil {
		return nil, fmt.Errorf("create OPAQUE server: %w", err)
	}

	if err := server.SetKeyMaterial(skm); err != nil {
		return nil, fmt.Errorf("set OPAQUE server key material: %w", err)
	}

	return &BytemareOpaqueServer{
		conf:     conf,
		serverID: []byte(serverIdentity),
		skm:      skm,
	}, nil
}

func (s *BytemareOpaqueServer) newServer() (*bytemare.Server, error) {
	server, err := s.conf.Server()
	if err != nil {
		return nil, err
	}

	if err := server.SetKeyMaterial(s.skm); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *BytemareOpaqueServer) RegisterStart(clientIdentity []byte, clientMessage []byte) ([]byte, []byte, error) {
	if len(clientIdentity) == 0 {
		return nil, nil, fmt.Errorf("empty client identity")
	}

	if len(clientMessage) == 0 {
		return nil, nil, fmt.Errorf("empty registration request")
	}

	server, err := s.newServer()
	if err != nil {
		return nil, nil, fmt.Errorf("create OPAQUE server: %w", err)
	}

	req, err := server.Deserialize.RegistrationRequest(clientMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("deserialize registration request: %w", err)
	}

	credentialIdentifier, err := NewFlowID()
	if err != nil {
		return nil, nil, fmt.Errorf("generate credential identifier: %w", err)
	}

	response, err := server.RegistrationResponse(req, []byte(credentialIdentifier), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create registration response: %w", err)
	}

	state := PendingOpaqueRegistrationState{
		CredentialIdentifier: credentialIdentifier,
		ClientIdentity:       string(clientIdentity),
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal registration state: %w", err)
	}

	return response.Serialize(), stateBytes, nil
}

func (s *BytemareOpaqueServer) RegisterFinish(serverState []byte, clientMessage []byte) ([]byte, error) {
	if len(serverState) == 0 {
		return nil, fmt.Errorf("empty registration state")
	}

	if len(clientMessage) == 0 {
		return nil, fmt.Errorf("empty registration record")
	}

	var state PendingOpaqueRegistrationState
	if err := json.Unmarshal(serverState, &state); err != nil {
		return nil, fmt.Errorf("decode registration state: %w", err)
	}

	if state.CredentialIdentifier == "" || state.ClientIdentity == "" {
		return nil, fmt.Errorf("invalid registration state")
	}

	server, err := s.newServer()
	if err != nil {
		return nil, fmt.Errorf("create OPAQUE server: %w", err)
	}

	record, err := server.Deserialize.RegistrationRecord(clientMessage)
	if err != nil {
		return nil, fmt.Errorf("deserialize registration record: %w", err)
	}

	stored := StoredOpaqueRecord{
		CredentialIdentifier: state.CredentialIdentifier,
		ClientIdentity:       state.ClientIdentity,
		RegistrationRecord:   base64.RawURLEncoding.EncodeToString(record.Serialize()),
	}

	storedBytes, err := json.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("marshal stored OPAQUE record: %w", err)
	}

	return storedBytes, err
}

func (s *BytemareOpaqueServer) LoginStart(storedRecord []byte, clientMessage []byte) ([]byte, []byte, error) {
	if len(storedRecord) == 0 {
		return nil, nil, fmt.Errorf("empty stored OPAQUE record")
	}

	if len(clientMessage) == 0 {
		return nil, nil, fmt.Errorf("empty KE1")
	}

	var stored StoredOpaqueRecord
	if err := json.Unmarshal(storedRecord, &stored); err != nil {
		return nil, nil, fmt.Errorf("decode stored OPAQUE record: %w", err)
	}

	if stored.CredentialIdentifier == "" ||
		stored.ClientIdentity == "" ||
		stored.RegistrationRecord == "" {
		return nil, nil, fmt.Errorf("invalid stored OPAQUE record")
	}

	recordBytes, err := base64.RawURLEncoding.DecodeString(stored.RegistrationRecord)
	if err != nil {
		return nil, nil, fmt.Errorf("decode registration record: %w", err)
	}

	server, err := s.newServer()
	if err != nil {
		return nil, nil, fmt.Errorf("create OPAQUE server: %w", err)
	}

	registrationRecord, err := server.Deserialize.RegistrationRecord(recordBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("deserialize registration record: %w", err)
	}

	ke1, err := server.Deserialize.KE1(clientMessage)
	if err != nil {
		return nil, nil, fmt.Errorf("deserialize KE1: %w", err)
	}

	clientRecord := &bytemare.ClientRecord{
		RegistrationRecord:   registrationRecord,
		CredentialIdentifier: []byte(stored.CredentialIdentifier),
		ClientIdentity:       []byte(stored.ClientIdentity),
	}

	ke2, output, err := server.GenerateKE2(ke1, clientRecord)
	if err != nil {
		return nil, nil, fmt.Errorf("generate KE2: %w", err)
	}

	state := PendingOpaqueLoginState{
		ExpectedClientMAC: base64.RawURLEncoding.EncodeToString(output.ClientMAC),
	}

	stateBytes, err := json.Marshal(state)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal login state: %w", err)
	}

	return ke2.Serialize(), stateBytes, nil
}

func (s *BytemareOpaqueServer) LoginFinish(serverState []byte, clientMessage []byte) error {
	if len(serverState) == 0 {
		return fmt.Errorf("empty login state")
	}

	if len(clientMessage) == 0 {
		return fmt.Errorf("empty KE3")
	}

	var state PendingOpaqueLoginState
	if err := json.Unmarshal(serverState, &state); err != nil {
		return fmt.Errorf("decode login state: %w", err)
	}

	expectedClientMAC, err := base64.RawURLEncoding.DecodeString(state.ExpectedClientMAC)
	if err != nil {
		return fmt.Errorf("decode expected client MAC: %w", err)
	}

	server, err := s.newServer()
	if err != nil {
		return fmt.Errorf("create OPAQUE server: %w", err)
	}

	ke3, err := server.Deserialize.KE3(clientMessage)
	if err != nil {
		return fmt.Errorf("deserialize KE3: %w", err)
	}

	if err := server.LoginFinish(ke3, expectedClientMAC); err != nil {
		return fmt.Errorf("finish OPAQUE login: %w", err)
	}

	return nil
}

func GenerateOpaqueServerKeyMaterial(serverIdentity string) ([]byte, error) {
	if serverIdentity == "" {
		return nil, fmt.Errorf("server identity is required")
	}

	conf := opaque.DefaultConfiguration()
	oprfSeed := conf.GenerateOPRFSeed()

	privateKey, publicKey := conf.KeyGen()
	if privateKey == nil || publicKey == nil {
		return nil, fmt.Errorf("generate OPAQUE server key pair")
	}

	skm := &opaque.ServerKeyMaterial{
		Identity:       []byte(serverIdentity),
		PrivateKey:     privateKey,
		PublicKeyBytes: publicKey.Encode(),
		OPRFGlobalSeed: oprfSeed,
	}

	encoded := skm.Encode()
	return encoded, nil
}