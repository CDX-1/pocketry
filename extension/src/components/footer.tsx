import { useEffect, useState } from "react";
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../components/ui/select";
import { Button } from "../components/ui/button";
import { PlusIcon, Settings2Icon, Trash2Icon } from "lucide-react";
import { Link } from "react-router-dom";
import { useServers } from "../components/context/servers-provider";
import { Popover, PopoverContent, PopoverTrigger } from "./ui/popover";
import { Label } from "./ui/label";
import { Input } from "./ui/input";

function Footer() {
    const {
        servers,
        selectedServerId,
        selectedServer,
        selectServer,
        updateServer,
        deleteServer,
    } = useServers();

    const [isPopoverOpen, setIsPopoverOpen] = useState(false);
    const [serverName, setServerName] = useState("");
    const [serverUrl, setServerUrl] = useState("");
    const [isSaving, setIsSaving] = useState(false);

    const hasServers = servers.length > 0;

    useEffect(() => {
        if (!selectedServer) {
            setServerName("");
            setServerUrl("");
            return;
        }

        setServerName(selectedServer.name);
        setServerUrl(selectedServer.url);
    }, [selectedServer]);

    async function handleSaveServer() {
        if (!selectedServer) {
            return;
        }

        const normalizedName = serverName.trim();
        const normalizedUrl = serverUrl.trim();

        if (!normalizedName || !normalizedUrl) {
            return;
        }

        setIsSaving(true);

        try {
            await updateServer(selectedServer.id, {
                name: normalizedName,
                url: normalizedUrl,
            });

            setIsPopoverOpen(false);
        } finally {
            setIsSaving(false);
        }
    }

    async function handleDeleteServer() {
        if (!selectedServer) {
            return;
        }

        const shouldDelete = window.confirm(
            `Delete "${selectedServer.name}"? This cannot be undone.`
        );

        if (!shouldDelete) {
            return;
        }

        setIsSaving(true);

        try {
            await deleteServer(selectedServer.id);
            setIsPopoverOpen(false);
        } finally {
            setIsSaving(false);
        }
    }

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

            <Popover open={isPopoverOpen} onOpenChange={setIsPopoverOpen}>
                <PopoverTrigger asChild disabled={!selectedServer}>
                    <Button size="sm" variant="ghost" className="h-9 shrink-0">
                        <Settings2Icon className="size-4" />
                    </Button>
                </PopoverTrigger>

                <PopoverContent align="end" className="w-[280px]">
                    <div className="grid gap-4">
                        <div className="space-y-2">
                            <h4 className="leading-none font-medium">
                                Edit the server
                            </h4>

                            <p className="text-sm text-muted-foreground">
                                You can edit the server name and URL.
                            </p>
                        </div>

                        <div className="grid gap-4">
                            <div className="grid grid-cols-3 items-center gap-4">
                                <Label htmlFor="serverName">Name</Label>

                                <Input
                                    id="serverName"
                                    value={serverName}
                                    onChange={(event) =>
                                        setServerName(event.target.value)
                                    }
                                    className="col-span-2 h-8"
                                />
                            </div>

                            <div className="grid grid-cols-3 items-center gap-4">
                                <Label htmlFor="serverUrl">URL</Label>

                                <Input
                                    id="serverUrl"
                                    type="url"
                                    value={serverUrl}
                                    onChange={(event) =>
                                        setServerUrl(event.target.value)
                                    }
                                    className="col-span-2 h-8"
                                />
                            </div>

                            <div className="flex items-center justify-end gap-2">
                                <Button
                                    type="button"
                                    size="sm"
                                    onClick={handleSaveServer}
                                    disabled={
                                        isSaving ||
                                        !serverName.trim() ||
                                        !serverUrl.trim()
                                    }
                                >
                                    {isSaving ? "Saving..." : "Save"}
                                </Button>

                                <Button
                                    type="button"
                                    size="sm"
                                    variant="destructive"
                                    onClick={handleDeleteServer}
                                    disabled={isSaving || !selectedServer}
                                >
                                    <Trash2Icon className="size-4" />
                                    Delete
                                </Button>
                            </div>
                        </div>
                    </div>
                </PopoverContent>
            </Popover>

            <Link to="/add-server" className="shrink-0">
                <Button size="sm" variant="ghost" className="h-9">
                    <PlusIcon className="size-4" />
                </Button>
            </Link>
        </div>
    );
}

export default Footer;