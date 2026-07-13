import type { LoginFinishResponse, LoginStartResponse, RegisterStartResponse } from "./auth-types";
import type { OpaqueClient, OpaqueLoginSession, OpaqueRegistrationSession } from "./opaque-client";

export class AuthService {
    private readonly baseURL: string;
    private readonly opaqueClient: OpaqueClient;
    private readonly serverIdentity: string;

    constructor(options: {
        serverURL: string,
        opaqueClient: OpaqueClient,
        serverIdentity?: string,
    }) {
        const serverURL = options.serverURL.trim();

        if (!serverURL) {
            throw new Error("server URL is required");
        }

        this.baseURL = serverURL.replace(/\/+$/, "");
        this.opaqueClient = options.opaqueClient;
        this.serverIdentity = options.serverIdentity?.trim() || "pocketry-server";
    }

    async register(
        usernameInput: string,
        password: string,
    ): Promise<{ serverStaticPublicKey: string }> {
        const username = this.normalizeUsername(usernameInput);

        if (!password) {
            throw new Error("password is required");
        }

        let session: OpaqueRegistrationSession | undefined;

        try {
            session = await this.opaqueClient.startRegistration(
                password,
                username,
                this.serverIdentity,
            );

            const startRes = await this.request<RegisterStartResponse>(
                "api/auth/register/start",
                {
                    method: 'POST',
                    body: JSON.stringify({
                        username,
                        client_message: session.clientMessage,
                    }),
                }
            )

            const opaqueResult = await session.finish(
                startRes.server_message,
            );

            await this.request(
                "api/auth/register/finish",
                {
                    method: "POST",
                    body: JSON.stringify({
                        registration_id: startRes.registration_id,
                        client_message: opaqueResult.registrationRecord,
                    })
                }
            )

            return {
                serverStaticPublicKey: opaqueResult.serverStaticPublicKey,
            }
        } finally {
            session?.dispose();
        }
    }

    async login(
        usernameInput: string,
        password: string,
    ): Promise<LoginFinishResponse & { serverStaticPublicKey: string }> {
        const username = this.normalizeUsername(usernameInput);

        if (!password) {
            throw new Error("password is required");
        }

        let session: OpaqueLoginSession | undefined;

        try {
            session = await this.opaqueClient.startLogin(
                password,
                username,
                this.serverIdentity,
            );

            const startRes = await this.request<LoginStartResponse>(
                "api/auth/login/start",
                {
                    method: "POST",
                    body: JSON.stringify({
                        username,
                        client_message: session.clientMessage,
                    }),
                }
            );

            const opaqueResult = await session.finish(startRes.server_message);

            const response = await this.request<LoginFinishResponse>(
                "api/auth/login/finish",
                {
                    method: "POST",
                    body: JSON.stringify({
                        login_id: startRes.login_id,
                        client_message: opaqueResult.finishLoginRequest,
                    }),
                }
            );

            return {
                ...response,
                serverStaticPublicKey: opaqueResult.serverStaticPublicKey,
            }
        } finally {
            session?.dispose();
        }
    }

    async getCurrentUser(
        accessToken: string,
    ): Promise<{ id: number; username: string }> {
        if (!accessToken.trim()) {
            throw new Error("access token is required");
        }

        return this.request("api/me", {
            method: "GET",
            headers: {
                Authorization: `Bearer ${accessToken}`,
            },
        });
    }

    private normalizeUsername(username: string): string {
        const normalized = username.trim().toLowerCase();

        if (!normalized) {
            throw new Error("username is required");
        }

        if (!/^[a-z0-9]{3,24}$/.test(normalized)) {
            throw new Error(
                "username must be between 3 and 24 alphanumeric characters",
            );
        }

        return normalized;
    }

    private async request<T = void>(
        path: string,
        options: RequestInit,
    ): Promise<T> {
        const response = await fetch(
            `${this.baseURL}/${path}`,
            {
                ...options,
                headers: {
                    "Content-Type": "application/json",
                    ...options.headers,
                },
            },
        );

        const responseText = await response.text();

        let body: unknown = null;

        if (responseText) {
            try {
                body = JSON.parse(responseText);
            } catch {
                body = responseText;
            }
        }

        if (!response.ok) {
            throw new Error(
                this.getErrorMessage(body, response.status),
            );
        }

        return body as T;
    }

    private getErrorMessage(
        body: unknown,
        status: number,
    ): string {
        if (
            typeof body === "object" &&
            body !== null &&
            "error" in body &&
            typeof body.error === "string"
        ) {
            return body.error;
        }

        if (typeof body === "string" && body.trim()) {
            return body;
        }

        return `request failed with status ${status}`;
    }
}