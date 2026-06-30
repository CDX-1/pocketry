import { useState } from "react";
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

const SERVERS = [
    {
        id: "server1",
        name: "Server 1",
        url: "http://localhost:3000",
    },
    {
        id: "server2",
        name: "Server 2",
        url: "http://localhost:4000",
    },
];

function Footer() {
    const [selectedServerId, setSelectedServerId] = useState<string>("");

    const selectedServer = SERVERS.find(
        (server) => server.id === selectedServerId
    );

    return (
        <div className="flex justify-between items-center gap-2 border-t border-border/5 bg-muted p-3">
            <Select value={selectedServerId} onValueChange={setSelectedServerId}>
                <SelectTrigger className="h-9 w-[280px]">
                    {selectedServer ? (
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
                        {SERVERS.map((server) => (
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

            <Button size="sm" variant="ghost" className="h-9">
                <Settings2Icon className="size-4" />
            </Button>

            <Link to="/add-server">
                <Button size="sm" variant="ghost" className="h-9">
                    <PlusIcon className="size-4" />
                </Button>
            </Link>
        </div>
    );
}

export default Footer;