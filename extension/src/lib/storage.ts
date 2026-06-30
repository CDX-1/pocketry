export type StoredServer = {
    id: string;
    name: string;
    url: string;
};

type StorageSchema = {
    servers: StoredServer[];
    selectedServerId?: string;
};

const DEFAULT_STORAGE: StorageSchema = {
    servers: [],
    selectedServerId: undefined,
};

const DEV_STORAGE_KEY = "pocketry-dev-storage";

// internal

function hasChromeStorage() {
    return typeof chrome !== "undefined" && !!chrome.storage?.local;
}

function getDevStorage(): StorageSchema {
    const raw = localStorage.getItem(DEV_STORAGE_KEY);

    if (!raw) {
        return DEFAULT_STORAGE;
    }

    try {
        return {
            ...DEFAULT_STORAGE,
            ...JSON.parse(raw),
        };
    } catch {
        return DEFAULT_STORAGE;
    }
}

function setDevStorage(data: Partial<StorageSchema>) {
    const currentStorage = getDevStorage();

    localStorage.setItem(
        DEV_STORAGE_KEY,
        JSON.stringify({
            ...currentStorage,
            ...data,
        })
    );
}

async function getStorage(): Promise<StorageSchema> {
    if (!hasChromeStorage()) {
        return getDevStorage();
    }

    const result: StorageSchema = await chrome.storage.local.get(DEFAULT_STORAGE);

    return {
        servers: result.servers ?? [],
        selectedServerId: result.selectedServerId,
    };
}

async function setStorage(data: Partial<StorageSchema>): Promise<void> {
    if (!hasChromeStorage()) {
        setDevStorage(data);
        return;
    }

    await chrome.storage.local.set(data);
}

// exported api

export async function getServers(): Promise<StoredServer[]> {
    const storage = await getStorage();
    return storage.servers;
}

export async function saveServers(servers: StoredServer[]): Promise<void> {
    await setStorage({ servers });
}

export async function saveServer(server: StoredServer): Promise<void> {
    const servers = await getServers();

    await saveServers([...servers, server]);
}

export async function getSelectedServerId(): Promise<string> {
    const storage = await getStorage();
    return storage.selectedServerId ?? "";
}

export async function saveSelectedServerId(serverId: string): Promise<void> {
    await setStorage({ selectedServerId: serverId });
}