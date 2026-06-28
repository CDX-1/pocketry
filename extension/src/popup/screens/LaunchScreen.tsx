import { Link } from "react-router-dom";
import { Button } from "../../components/ui/button";

function LaunchScreen() {
    return (
        <div className="flex h-full flex-col justify-end mb-2">
            <div className="flex flex-col gap-y-1">
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
    );
}

export default LaunchScreen;