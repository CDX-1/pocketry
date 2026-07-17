import { Link, useNavigate } from "react-router-dom";
import { Button } from "../../components/ui/button";
import { useEffect } from "react";
import { useServers } from "../../components/context/servers-provider";
import { getServerSession } from "../../lib/session";

function LaunchScreen() {
    const navigate = useNavigate();
    const {
        servers,
        selectedServerId,
        selectedServer,
        selectServer,
        isLoading,
    } = useServers();

    // determine the first screen to show
    useEffect(() => {
        if (isLoading) {
            return;
        }

        let cancelled = false;

        async function determineInitialScreen() {
            try {
                if (servers.length === 0) {
                    navigate("/add-server", {
                        replace: true,
                    });
                    return;
                }

                let activeServer = selectedServer;

                if (!activeServer) {
                    const firstServer = servers[0];

                    if (!firstServer) {
                        navigate("/add-server", {
                            replace: true,
                        });
                        return;
                    }

                    await selectServer(firstServer.id);
                    if (cancelled) return;

                    activeServer = firstServer;
                }

                const session = await getServerSession(activeServer.id);
                if (cancelled) return;

                if (session) {
                    navigate("/dashboard", {
                        replace: true,
                    });
                    return;
                }

                navigate("/login", {
                    replace: true,
                });
            } catch (error) {
                console.error("Failed to determine initial screen:", error);
                if (!cancelled) {
                    navigate("/login", {
                        replace: true,
                    });
                }
            }
        }

        void determineInitialScreen();

        return () => {
            cancelled = true;
        };
    }, [
        isLoading, servers, selectedServerId,
        selectedServer, selectServer, navigate,
    ]);

    if (isLoading) {
        return (
            <div className="flex h-full items-center justify-center">
                <p className="text-sm text-muted-foreground">
                    Loading Pocketry...
                </p>
            </div>
        )
    }

    return (
        <div className="flex flex-col h-full pb-8">
            <p className="mt-40 text-center">
                <span className="text-lg font-semibold">Welcome to Pocketry.</span>
                <br />
                The self-hostable, zero-knowledge, PQC-encrypted password manager.
                <br />
                <br />
                <span className="text-sm text-muted-foreground">Press the 'add server' button to get started.</span>
            </p>

            <div className="flex h-full flex-col justify-end">
                <div className="flex flex-col gap-y-1 mt-4">
                    <Button size="lg" asChild className="flex w-full items-center">
                        <Link to="/add-server">Add server</Link>
                    </Button>

                    <Button
                        asChild
                        size="sm"
                        variant="link"
                        className="flex w-full items-center text-xs text-muted-foreground"
                    >
                        <a
                            href="https://github.com/CDX-1/pocketry"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            Host your own server
                        </a>
                    </Button>
                </div>
            </div>
        </div>
    );
}

export default LaunchScreen;