export interface OpaqueRegistrationSession {
    readonly clientMessage: Uint8Array;

    finish(serverMessage: Uint8Array): Promise<Uint8Array>;

    dispose(): void;
}

export interface OpaqueLoginSession {
    readonly clientMessage: Uint8Array;

    finish(serverMessage: Uint8Array): Promise<Uint8Array>;

    dispose(): void;
}

export interface OpaqueClient {
    startRegistration(
        password: Uint8Array,
        clientIdentity: Uint8Array,
        serverIdentity: Uint8Array,
    ): Promise<OpaqueRegistrationSession>;

    startLogin(
        password: Uint8Array,
        clientIdentity: Uint8Array,
        serverIdentity: Uint8Array,
    ): Promise<OpaqueLoginSession>;
}