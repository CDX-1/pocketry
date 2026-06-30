import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
    type ReactNode,
} from "react";

import {
    getSelectedServerId,
    getServers,
    saveSelectedServerId,
    saveServers,
    type StoredServer,
} from "../../lib/storage";

type ServersContextValue = {
    servers: StoredServer[];
    selectedServerId: string;
    selectedServer?: StoredServer;
    isLoading: boolean;
    addServer: (server: Omit<StoredServer, "id">) => Promise<StoredServer>;
    selectServer: (serverId: string) => Promise<void>;
    refreshServers: () => Promise<void>;
};

const ServersContext = createContext<ServersContextValue | null>(null);

export function ServersProvider({ children }: { children: ReactNode }) {
    const [servers, setServers] = useState<StoredServer[]>([]);
    const [selectedServerId, setSelectedServerId] = useState("");
    const [isLoading, setIsLoading] = useState(true);

    const refreshServers = useCallback(async () => {
        const [storedServers, storedSelectedServerId] = await Promise.all([
            getServers(),
            getSelectedServerId(),
        ]);

        setServers(storedServers);
        setSelectedServerId(storedSelectedServerId ?? "");
    }, []);

    useEffect(() => {
        async function load() {
            try {
                await refreshServers();
            } finally {
                setIsLoading(false);
            }
        }

        void load();
    }, [refreshServers]);

    const selectedServer = useMemo(
        () => servers.find((server) => server.id === selectedServerId),
        [servers, selectedServerId]
    );

    const selectServer = useCallback(async (serverId: string) => {
        setSelectedServerId(serverId);
        await saveSelectedServerId(serverId);
    }, []);

    const addServer = useCallback(
        async (serverInput: Omit<StoredServer, "id">) => {
            const newServer: StoredServer = {
                id: crypto.randomUUID(),
                ...serverInput,
            };

            const currentServers = await getServers();
            const nextServers = [...currentServers, newServer];

            await saveServers(nextServers);

            setServers(nextServers);

            const currentSelectedServerId = await getSelectedServerId();

            if (!currentSelectedServerId) {
                await saveSelectedServerId(newServer.id);
                setSelectedServerId(newServer.id);
            }

            return newServer;
        },
        []
    );

    const value = useMemo(
        () => ({
            servers,
            selectedServerId,
            selectedServer,
            isLoading,
            addServer,
            selectServer,
            refreshServers,
        }),
        [
            servers,
            selectedServerId,
            selectedServer,
            isLoading,
            addServer,
            selectServer,
            refreshServers,
        ]
    );

    return (
        <ServersContext.Provider value={value}>
            {children}
        </ServersContext.Provider>
    );
}

export function useServers() {
    const context = useContext(ServersContext);

    if (!context) {
        throw new Error("useServers must be used inside ServersProvider");
    }

    return context;
}