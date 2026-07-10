package auth

type OpaqueServer interface {
	RegisterStart(clientIdentity []byte, clientMessage []byte) (serverMessage []byte, serverState []byte, err error)
	RegisterFinish(serverState []byte, clientMessage []byte) (storedRecord []byte, err error)

	LoginStart(storedRecord []byte, clientMessage []byte) (serverMessage []byte, serverState []byte, err error)
	LoginFinish(serverState []byte, clientMessage []byte) error
}