package auth

import (
	bytemare "github.com/bytemare/opaque"
)

const defaultOpaqueServerIdentity = "pocketry-server"

type BytemareOpaqueServer struct {
	conf	 *bytemare.Configuration
	serverID []BytemareOpaqueServer
	skm		 *bytemare.ServerKeyMaterial
}
