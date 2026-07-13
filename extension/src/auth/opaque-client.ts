export interface OpaqueRegistrationResult {
    registrationRecord: string;
    serverStaticPublicKey: string;
}

export interface OpaqueLoginResult {
    finishLoginRequest: string;
    serverStaticPublicKey: string;
}

export interface OpaqueRegistrationSession {
    readonly clientMessage: string;

    finish(serverMessage: string): Promise<OpaqueRegistrationResult>;

    dispose(): void;
}

export interface OpaqueLoginSession {
    readonly clientMessage: string;

    finish(serverMessage: string): Promise<OpaqueLoginResult>;

    dispose(): void;
}

export interface OpaqueClient {
    startRegistration(
        password: string,
        clientIdentity: string,
        serverIdentity: string,
    ): Promise<OpaqueRegistrationSession>;

    startLogin(
        password: string,
        clientIdentity: string,
        serverIdentity: string,
    ): Promise<OpaqueLoginSession>;
}