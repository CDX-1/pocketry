package auth

import (
	"encoding/json"
	"strings"
	"testing"
)

func mustMarshalJSON(t *testing.T, value any) []byte {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal test value: %v", err)
	}

	return encoded
}


// ensures opaque bytemare rejects if skm is missing
func TestNewBytemareOpaqueServerRejectsMissingKeyMaterial(t *testing.T) {
	server, err := NewBytemareOpaqueServer("test-server", nil)

	if err == nil {
		t.Fatal("expected err, got nil")
	}

	if server != nil {
		t.Fatal("expected server to be nil")
	}

	const want = "OPAQUE server key material is required"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects invalid skm
func TestNewBytemareOpaqueServerRejectsInvalidKeyMaterial(t *testing.T) {
	server, err := NewBytemareOpaqueServer(
		"test-server",
		[]byte("invalid-key-material"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if server != nil {
		t.Fatal("expected server to be nil")
	}

	if !strings.Contains(
		err.Error(),
		"decode OPAQUE server key material",
	) {
		t.Fatalf(
			"expected key-material decoding error, got %v",
			err,
		)
	}
}

// ensures opaque bytemare rejects empty client identity during registration start
func TestRegisterStartRejectsEmptyClientIdentity(t *testing.T) {
	server := &BytemareOpaqueServer{}

	response, state, err := server.RegisterStart(
		nil,
		[]byte("registration-request"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected nil state, got %v", state)
	}

	const want = "empty client identity"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects empty registration request during registration start
func TestRegisterStartRejectsEmptyRegistrationRequest(t *testing.T) {
	server := &BytemareOpaqueServer{}

	response, state, err := server.RegisterStart(
		[]byte("username"),
		nil,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected bil state, got %v", state)
	}

	const want = "empty registration request"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects empty server state during registration finish
func TestRegisterFinishRejectsEmptyServerState(t *testing.T) {
	server := &BytemareOpaqueServer{}

	record, err := server.RegisterFinish(
		nil,
		[]byte("registration-record"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if record != nil {
		t.Fatalf("expected nil record, got %v", record)
	}

	const want = "empty registration state"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects invalid server state during registration finish
func TestRegisterFinishRejectsEmptyRegistrationRecord(t *testing.T) {
	server := &BytemareOpaqueServer{}

	state := mustMarshalJSON(t, PendingOpaqueRegistrationState{
		CredentialIdentifier: "credential-in",
		ClientIdentity:       "username",
	})

	record, err := server.RegisterFinish(state, nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if record != nil {
		t.Fatalf("expected nil record, got %v", record)
	}

	const want = "empty registration record"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects malformed state JSON during registration finish
func TestRegisterFinishRejectsMalformedStateJSON(t *testing.T) {
	server := &BytemareOpaqueServer{}

	record, err := server.RegisterFinish(
		[]byte("{invalid-json}"),
		[]byte("registration-record"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if record != nil {
		t.Fatalf("expected nil record, got %v", record)
	}

	if !strings.Contains(err.Error(), "decode registration state") {
		t.Fatalf("expected decode registration state error, got %v", err)
	}
}

// ensures opaque bytemare rejects incomplete registration state during registration finish
func TestRegisterFinishRejectsIncompleteState(t *testing.T) {
	tests := []struct {
		name  string
		state PendingOpaqueRegistrationState
	}{
		{
			name: "missing credential identifier",
			state: PendingOpaqueRegistrationState{
				ClientIdentity: "username",
			},
		},
		{
			name: "missing client identity",
			state: PendingOpaqueRegistrationState{
				CredentialIdentifier: "credential-id",
			},
		},
		{
			name: "both fields missing",
			state: PendingOpaqueRegistrationState{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &BytemareOpaqueServer{}
			state := mustMarshalJSON(t, tt.state)

			record, err := server.RegisterFinish(
				state,
				[]byte("registration-record"),
			)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if record != nil {
				t.Fatalf("expected nil record, got %v", record)
			}

			const want = "invalid registration state"

			if err.Error() != want {
				t.Fatalf(
					"expected error %q, got %q",
					want,
					err.Error(),
				)
			}
		})
	}
}

// ensures opaque bytemare rejects empty stored record during login start
func TestLoginStartRejectsEmptyStoredRecord(t *testing.T) {
	server := &BytemareOpaqueServer{}

	response, state, err := server.LoginStart(
		nil,
		[]byte("ke1"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected nil state, got %v", state)
	}

	const want = "empty stored OPAQUE record"

	if err.Error() != want {
		t.Fatalf(
			"expected error %q, got %q",
			want,
			err.Error(),
		)
	}
}

// ensures opaque bytemare rejects empty KE1 during login start
func TestLoginStartRejectsEmptyKE1(t *testing.T) {
	server := &BytemareOpaqueServer{}

	response, state, err := server.LoginStart(
		[]byte(`{"registration_record":"im so tired of writing tests"}`),
		nil,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected nil state, got %v", state)
	}

	const want = "empty KE1"

	if err.Error() != want {
		t.Fatalf(
			"expected error %q, got %q",
			want,
			err.Error(),
		)
	}
}

// ensures opaque bytemare rejects malformed stored record json during login start
func TestLoginStartRejectsMalformedStoredRecordJSON(t *testing.T) {
	server := &BytemareOpaqueServer{}

	response, state, err := server.LoginStart(
		[]byte("{invalid-json}"),
		[]byte("ke1"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected nil state, got %v", state)
	}

	if !strings.Contains(err.Error(), "decode stored OPAQUE record") {
		t.Fatalf(
			"expected decode stored OPAQUE record error, got %v",
			err,
		)
	}
}

// ensures opaque bytemare rejects invalid registration record base64 during login start
func TestLoginStartRejectsInvalidRegistrationRecordBase64(t *testing.T) {
	server := &BytemareOpaqueServer{}

	storedRecord := mustMarshalJSON(t, StoredOpaqueRecord{
		CredentialIdentifier: "credential-id",
		ClientIdentity: "username",
		RegistrationRecord: "not!base64",
	})

	response, state, err := server.LoginStart(
		storedRecord,
		[]byte("ke1"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if response != nil {
		t.Fatalf("expected nil response, got %v", response)
	}

	if state != nil {
		t.Fatalf("expected nil state, got %v", state)
	}

	if !strings.Contains(err.Error(), "decode registration record") {
		t.Fatalf(
			"expected decode registration record error, got %v",
			err,
		)
	}
}

// ensures opaque bytemare rejects empty login state during login finish
func TestLoginFinishRejectsEmptyServerState(t *testing.T) {
	server := &BytemareOpaqueServer{}

	err := server.LoginFinish(nil, []byte("ke3"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	const want = "empty login state"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects empty ke3 during login finish
func TestLoginFinishRejectsEmptyKE3(t *testing.T) {
	server := &BytemareOpaqueServer{}

	state := mustMarshalJSON(t, PendingOpaqueLoginState{
		ExpectedClientMAC: "YWJj",
	})

	err := server.LoginFinish(state, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	const want = "empty KE3"

	if err.Error() != want {
		t.Fatalf("expected error %q, got %q", want, err.Error())
	}
}

// ensures opaque bytemare rejects malformed login state json during login finish
func TestLoginFinishRejectsMalformedStateJSON(t *testing.T) {
	server := &BytemareOpaqueServer{}

	err := server.LoginFinish(
		[]byte("{invalid-json"),
		[]byte("ke3"),
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "decode login state") {
		t.Fatalf(
			"expected login-state decoding error, got %q",
			err.Error(),
		)
	}
}

// ensures opaque bytemare rejects invalid client MAC base64 during login finish
func TestLoginFinishRejectsInvalidClientMACBase64(t *testing.T) {
	server := &BytemareOpaqueServer{}

	state := mustMarshalJSON(t, PendingOpaqueLoginState{
		ExpectedClientMAC: "not!base64",
	})

	err := server.LoginFinish(state, []byte("ke3"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"decode expected client MAC",
	) {
		t.Fatalf(
			"expected decode expected client MAC error, got %v",
			err,
		)
	}
}


// ensures opaque bytemare can round-trip pending registration state to json
func TestPendingOpaqueRegistrationStateJSONRoundTrip(t *testing.T) {
	want := PendingOpaqueRegistrationState{
		CredentialIdentifier: "credential-id",
		ClientIdentity:		  "username",
	}

	encoded := mustMarshalJSON(t, want)

	var got PendingOpaqueRegistrationState
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal registration state: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected state: got %+v, want %+v", got, want)
	}
}

// ensures opaque bytemare can round-trip pending login state to json
func TestPendingOpaqueLoginStateJSONRoundTrip(t *testing.T) {
	want := PendingOpaqueLoginState{
		ExpectedClientMAC: "ZXhwZWN0ZWQtbWFj",
	}

	encoded := mustMarshalJSON(t, want)

	var got PendingOpaqueLoginState
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal login state: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected state: got %+v, want %+v", got, want)
	}
}

// ensures opaque bytemare can round-trip stored record to json
func TestStoredOpaqueRecordJSONRoundTrip(t *testing.T) {
	want := StoredOpaqueRecord{
		CredentialIdentifier: "credential-id",
		ClientIdentity:       "testuser",
		RegistrationRecord:   "cmVnaXN0cmF0aW9uLXJlY29yZA",
	}

	encoded := mustMarshalJSON(t, want)

	var got StoredOpaqueRecord
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatalf("unmarshal stored record: %v", err)
	}

	if got != want {
		t.Fatalf("unexpected record: got %+v, want %+v", got, want)
	}
}