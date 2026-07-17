import * as opaque from "@serenity-kit/opaque";
import type { OpaqueClient, OpaqueLoginResult, OpaqueLoginSession, OpaqueRegistrationResult, OpaqueRegistrationSession } from "./opaque-client";

export class SerenityOpaqueClient implements OpaqueClient {
    private readyPromise: Promise<void> | undefined;

    async startRegistration(
        password: string,
        clientIdentity: string,
        serverIdentity: string,
    ): Promise<OpaqueRegistrationSession> {
        await this.ensureReady();

        const { clientRegistrationState, registrationRequest } = opaque.client.startRegistration({
            password,
        });

        let disposed = false;

        return {
            clientMessage: registrationRequest,

            async finish(serverMessage: string): Promise<OpaqueRegistrationResult> {
                if (disposed) {
                    throw new Error("OPAQUE registration session has been disposed");
                }

                const result = opaque.client.finishRegistration({
                    clientRegistrationState,
                    registrationResponse: serverMessage,
                    password,
                    identifiers: {
                        client: clientIdentity,
                        server: serverIdentity,
                    },
                });

                return {
                    registrationRecord: result.registrationRecord,
                    serverStaticPublicKey: result.serverStaticPublicKey,
                }
            },

            dispose(): void {
                disposed = true;
            }
        }
    }

    async startLogin(
        password: string,
        clientIdentity: string,
        serverIdentity: string,
    ): Promise<OpaqueLoginSession> {
        await this.ensureReady();

        const { clientLoginState, startLoginRequest } = opaque.client.startLogin({
            password,
        });

        let disposed = false;

        return {
            clientMessage: startLoginRequest,

            async finish(serverMessage: string): Promise<OpaqueLoginResult> {
                if (disposed) {
                    throw new Error("OPAQUE login session has been disposed");
                }

                const result = opaque.client.finishLogin({
                    clientLoginState,
                    loginResponse: serverMessage,
                    password,
                    identifiers: {
                        client: clientIdentity,
                        server: serverIdentity,
                    },
                });

                if (!result) {
                    throw new Error("invalid username or password");
                }

                return {
                    finishLoginRequest: result.finishLoginRequest,
                    serverStaticPublicKey: result.serverStaticPublicKey,
                }
            },

            dispose(): void {
                disposed = true;
            }
        }
    }

    private async ensureReady(): Promise<void> {
        this.readyPromise ??= opaque.ready;
        return this.readyPromise;
    }
}