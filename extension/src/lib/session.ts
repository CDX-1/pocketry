type SessionSchema = {
    sessions: Record<string, StoredSession>;
}

export type StoredSession = {
    accessToken: string;
    expiresAt: number;
}

const DEFAULT_SESSION_STORAGE: SessionSchema = {
    sessions: {},
};

function hasChromeSessionStorage(): boolean {
    return (
        typeof chrome !== "undefined" &&
        !!chrome.storage?.session
    );
}

async function getSessionStorage(): Promise<SessionSchema> {
    if (!hasChromeSessionStorage()) {
        return {
            sessions: {},
        };
    }

    const result = await chrome.storage.session.get(DEFAULT_SESSION_STORAGE);

    return {
        sessions: (result.sessions as Record<string, StoredSession>) ?? {},
    };
}

async function setSessionStorage(data: Partial<SessionSchema>): Promise<void> {
    if (!hasChromeSessionStorage) {
        return;
    }

    await chrome.storage.session.set(data);
}

export async function saveServerSession(serverId: string, session: StoredSession) {
    const storage = await getSessionStorage();

    await setSessionStorage({
        sessions: {
            ...storage.sessions,
            [serverId]: session,
        },
    });
}

export async function getServerSession(serverId: string): Promise<StoredSession | undefined> {
    const storage = await getSessionStorage();
    const session = storage.sessions[serverId];

    if (!session) {
        return undefined;
    }

    if (Date.now() >= session.expiresAt) {
        await clearServerSession(serverId);
        return undefined;
    }

    return session;
}

export async function clearServerSession(serverId: string): Promise<void> {
    const storage = await getSessionStorage();
    const sessions = { ...storage.sessions };

    delete sessions[serverId];

    await setSessionStorage({ sessions });
}