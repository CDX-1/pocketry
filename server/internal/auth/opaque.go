package auth

type OpaqueServer interface {
	RegisterStart(clientMessage []byte) (serverMessage []byte, serverState []byte, err error)
	RegisterFinish(serverState []byte, clientMessage []byte) (registrationRecord []byte, err error)

	LoginStart(registrationRecord []byte, clientMessage []byte) (serverMessage []byte, serverState []byte, err error)
	LoginFinish(serverState []byte, clientMessage []byte) error
}