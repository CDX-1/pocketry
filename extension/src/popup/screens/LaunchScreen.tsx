import { Link } from "react-router-dom";
import { Button } from "../../components/ui/button";

function LaunchScreen() {
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