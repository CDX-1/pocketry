import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../components/ui/select";
import { Button } from "../components/ui/button";
import { PlusIcon, Settings2Icon } from "lucide-react";
import { Link } from "react-router-dom";
import { useServers } from "../components/context/servers-provider";

function Footer() {
    const {
        servers,
        selectedServerId,
        selectedServer,
        selectServer,
    } = useServers();

    const hasServers = servers.length > 0;

    return (
        <div className="flex items-center justify-between gap-2 border-t border-border/5 bg-muted p-3">
            <Select
                value={selectedServerId}
                onValueChange={selectServer}
                disabled={!hasServers}
            >
                <SelectTrigger className="h-9 w-full min-w-0">
                    {!hasServers ? (
                        <span className="truncate text-sm text-muted-foreground">
                            No servers, + to add a server
                        </span>
                    ) : selectedServer ? (
                        <span className="flex w-full min-w-0 items-center justify-between gap-3 pr-1">
                            <span className="max-w-[110px] truncate text-sm font-medium text-foreground">
                                {selectedServer.name}
                            </span>

                            <span className="min-w-0 flex-1 truncate text-right text-xs text-muted-foreground">
                                {selectedServer.url}
                            </span>
                        </span>
                    ) : (
                        <SelectValue placeholder="Select a server..." />
                    )}
                </SelectTrigger>

                <SelectContent
                    position="popper"
                    align="start"
                    side="bottom"
                    sideOffset={6}
                    className="w-[280px]"
                >
                    <SelectGroup>
                        {servers.map((server) => (
                            <SelectItem
                                key={server.id}
                                value={server.id}
                                textValue={`${server.name} ${server.url}`}
                            >
                                <span className="flex w-full min-w-0 items-center justify-between gap-3 pr-5">
                                    <span className="max-w-[95px] truncate text-sm font-medium">
                                        {server.name}
                                    </span>

                                    <span className="min-w-0 flex-1 truncate text-right text-xs text-muted-foreground">
                                        {server.url}
                                    </span>
                                </span>
                            </SelectItem>
                        ))}
                    </SelectGroup>
                </SelectContent>
            </Select>

            <Button size="sm" variant="ghost" className="h-9 shrink-0">
                <Settings2Icon className="size-4" />
            </Button>

            <Link to="/add-server" className="shrink-0">
                <Button size="sm" variant="ghost" className="h-9">
                    <PlusIcon className="size-4" />
                </Button>
            </Link>
        </div>
    );
}

export default Footer;