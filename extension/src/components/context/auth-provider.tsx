import { createContext, useCallback, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { clearServerSession, getServerSession, saveServerSession, type StoredSession } from "../../lib/session";
import { useServers } from "./servers-provider";
import { AuthService } from "../../auth/auth-service";
import { SerenityOpaqueClient } from "../../auth/serenity-opaque-client";
import { useNotifications } from "./notification-provider";

type AuthStatus = "loading" | "authenticated" | "unauthenticated";

type CurrentUser = {
    id: number;
    username: string;
}

type AuthContextValue = {
    status: AuthStatus;
    session?: StoredSession;
    currentUser?: CurrentUser;

    register: (username: string, password: string) => Promise<void>;
    login: (username: string, password: string) => Promise<void>;

    logout: () => Promise<void>;
    refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
    const { selectedServer, updateServer } = useServers();
    const { notify } = useNotifications();

    const [status, setStatus] = useState<AuthStatus>("loading");
    const [session, setSession] = useState<StoredSession>();
    const [currentUser, setCurrentUser] = useState<CurrentUser>();

    const createAuthService = useCallback(() => {
        if (!selectedServer) {
            throw new Error("no server selected");
        }

        return new AuthService({
            serverURL: selectedServer.url,
            opaqueClient: new SerenityOpaqueClient(),
            serverIdentity: "pocketry-server",
        });
    }, [selectedServer]);

    const refreshSession = useCallback(async () => {
        setStatus("loading");

        if (!selectedServer) {
            setSession(undefined);
            setCurrentUser(undefined);
            setStatus("unauthenticated");
            return;
        }

        try {
            const storedSession = await getServerSession(selectedServer.id);

            if (!storedSession) {
                setSession(undefined);
                setCurrentUser(undefined);
                setStatus("unauthenticated");
                return;
            }

            const authService = createAuthService();
            const user = await authService.getCurrentUser(storedSession.accessToken);

            setSession(storedSession);
            setCurrentUser(user);
            setStatus("authenticated");
        } catch {
            await clearServerSession(selectedServer.id);

            setSession(undefined);
            setCurrentUser(undefined);
            setStatus("unauthenticated");
        }
    }, [selectedServer, createAuthService]);

    const register = useCallback(async (username: string, password: string) => {
        if (!selectedServer) {
            throw new Error("no server selected");
        }

        const authService = createAuthService();

        const result = await authService.register(username, password);

        await updateServer(selectedServer.id, {
            username: username.trim().toLowerCase(),
            serverStaticPublicKey: result.serverStaticPublicKey,
        });
    }, [selectedServer, createAuthService, updateServer]);

    const login = useCallback(async (username: string, password: string) => {
        if (!selectedServer) {
            throw new Error("no server selected");
        }

        const authService = createAuthService();

        const result = await authService.login(username, password);
        const pinnedKey = selectedServer.serverStaticPublicKey;

        if (pinnedKey && pinnedKey !== result.serverStaticPublicKey) {
            throw new Error("The server security key has changed. Login failed.");
        }

        const nextSession: StoredSession = {
            accessToken: result.access_token,
            expiresAt: Date.now() + result.expires_in * 1000,
        };

        await updateServer(selectedServer.id, {
            username: username.trim().toLowerCase(),
            serverStaticPublicKey: result.serverStaticPublicKey,
        });

        await saveServerSession(selectedServer.id, nextSession);

        const user = await authService.getCurrentUser(result.access_token);

        setSession(nextSession);
        setCurrentUser(user);
        setStatus("authenticated");
    }, [selectedServer, createAuthService, updateServer]);

    const clearAuthentication = useCallback(async () => {
        if (selectedServer) {
            await clearServerSession(selectedServer.id);
        }

        setSession(undefined);
        setCurrentUser(undefined);
        setStatus("unauthenticated");
    }, [selectedServer]);

    const logout = useCallback(async () => {
        await clearAuthentication();
    }, [clearAuthentication]);

    const expireSession = useCallback(async () => {
        await clearAuthentication();

        notify({
            title: "Session expired",
            description: "Your session has expired. Please log in again.",
            variant: "destructive",
            ttl: 0,
        });
    }, [clearAuthentication, notify]);

    useEffect(() => {
        void refreshSession();
    }, [refreshSession]);

    useEffect(() => {
        if (status !== "authenticated" || !session) return;

        const remainingTime = session.expiresAt - Date.now();
        if (remainingTime <= 0) {
            void expireSession();
            return;
        }

        const timeoutId = window.setTimeout(() => {
            void expireSession();
        }, remainingTime);

        return () => {
            window.clearTimeout(timeoutId);
        }
    }, [status, session, expireSession]);

    const value = useMemo<AuthContextValue>(
        () => ({
            status,
            session,
            currentUser,
            register,
            login,
            logout,
            refreshSession,
        }),
        [
            status,
            session,
            currentUser,
            register,
            login,
            logout,
            refreshSession,
        ]
    );

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    )
}

export function useAuth(): AuthContextValue {
    const context = useContext(AuthContext);

    if (!context) {
        throw new Error("useAuth must be used inside AuthProvider");
    }

    return context;
}