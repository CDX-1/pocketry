import { base64URLToBytes, bytesToBase64URL } from "../crypto/base64url";
import type { LoginFinishResponse, LoginStartResponse, RegisterStartResponse } from "./auth-types";
import type { OpaqueClient, OpaqueLoginSession, OpaqueRegistrationSession } from "./opaque-client";

const textEncoder = new TextEncoder();

export class AuthService {
    private readonly baseURL: string;
    private readonly opaqueClient: OpaqueClient;
    private readonly serverIdentity: Uint8Array;

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
        this.serverIdentity = textEncoder.encode(
            options.serverIdentity ?? "pocketry-server"
        );
    }

    async register(
        usernameInput: string,
        passwordInput: string,
    ): Promise<void> {
        const username = this.normalizeUsername(usernameInput);

        if (!passwordInput) {
            throw new Error("password is required");
        }

        const usernameBytes = textEncoder.encode(username);
        const passwordBytes = textEncoder.encode(passwordInput);

        let session: OpaqueRegistrationSession | undefined;

        try {
            session = await this.opaqueClient.startRegistration(
                passwordBytes,
                usernameBytes,
                this.serverIdentity,
            );

            const startRes = await this.request<RegisterStartResponse>(
                "api/auth/register/start",
                {
                    method: 'POST',
                    body: JSON.stringify({
                        username,
                        client_message: bytesToBase64URL(session.clientMessage),
                    }),
                }
            )

            const registrationRecord = await session.finish(
                base64URLToBytes(startRes.server_message),
            );

            await this.request(
                "api/auth/register/finish",
                {
                    method: "POST",
                    body: JSON.stringify({
                        registration_id: startRes.registration_id,
                        client_message: bytesToBase64URL(registrationRecord),
                    })
                }
            )
        } finally {
            passwordBytes.fill(0);
            session?.dispose();
        }
    }

    async login(
        usernameInput: string,
        passwordInput: string,
    ): Promise<LoginFinishResponse> {
        const username = this.normalizeUsername(usernameInput);

        if (!passwordInput) {
            throw new Error("password is required");
        }

        const usernameBytes = textEncoder.encode(username);
        const passwordBytes = textEncoder.encode(passwordInput);

        let session: OpaqueLoginSession | undefined;

        try {
            session = await this.opaqueClient.startLogin(
                passwordBytes,
                usernameBytes,
                this.serverIdentity,
            );

            const startRes = await this.request<LoginStartResponse>(
                "api/auth/login/start",
                {
                    method: "POST",
                    body: JSON.stringify({
                        username,
                        client_message: bytesToBase64URL(session.clientMessage),
                    }),
                }
            );

            const ke3 = await session.finish(
                base64URLToBytes(startRes.server_message),
            );

            return await this.request<LoginFinishResponse>(
                "api/auth/login/finish",
                {
                    method: "POST",
                    body: JSON.stringify({
                        login_id: startRes.login_id,
                        client_message: bytesToBase64URL(ke3),
                    }),
                }
            )
        } finally {
            passwordBytes.fill(0);
            session?.dispose();
        }
    }

    async getCurrentUser(
        accessToken: string,
    ): Promise<{ id: number, username: string }> {
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

        const body: unknown = await response
            .json()
            .catch(() => null);

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

        return `request failed with status ${status}`;
    }
}