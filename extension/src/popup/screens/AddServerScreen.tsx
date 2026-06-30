import { type FormEvent, useState } from "react";
import { Button } from "../../components/ui/button";
import {
    Field,
    FieldDescription,
    FieldGroup,
    FieldLabel,
    FieldLegend,
    FieldSet,
} from "../../components/ui/field";
import { Input } from "../../components/ui/input";

function AddServerScreen() {
    const [serverName, setServerName] = useState("");
    const [serverUrl, setServerUrl] = useState("");

    function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const normalizedServerName = serverName.trim();
        const normalizedServerUrl = serverUrl.trim();

        console.log("Server:", {
            name: normalizedServerName,
            url: normalizedServerUrl,
        });
    }

    return (
        <form className="flex h-full flex-col justify-end pb-8" onSubmit={handleSubmit}>
            <FieldSet>
                <FieldLegend>Add new server</FieldLegend>

                <FieldDescription>
                    Add a Pocketry server you want to connect to.
                </FieldDescription>

                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="server-name">Server name</FieldLabel>

                        <FieldDescription className="text-xs">
                            A friendly name for this server.
                        </FieldDescription>

                        <Input
                            id="server-name"
                            type="text"
                            placeholder="Self-hosted server"
                            value={serverName}
                            onChange={(event) => setServerName(event.target.value)}
                            required
                        />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="server-url">Server URL</FieldLabel>

                        <FieldDescription className="text-xs">
                            The server's URL.
                        </FieldDescription>

                        <Input
                            id="server-url"
                            type="url"
                            placeholder="https://example.com"
                            value={serverUrl}
                            onChange={(event) => setServerUrl(event.target.value)}
                            required
                        />
                    </Field>

                    <Button size="lg" type="submit">Add server</Button>
                </FieldGroup>
            </FieldSet>
        </form>
    );
}

export default AddServerScreen;