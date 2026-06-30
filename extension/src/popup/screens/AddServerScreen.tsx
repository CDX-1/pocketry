import { type FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

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
import { useServers } from "../../components/context/servers-provider";

function AddServerScreen() {
    const navigate = useNavigate();
    const { addServer } = useServers();

    const [serverName, setServerName] = useState("");
    const [serverUrl, setServerUrl] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const normalizedServerName = serverName.trim();
        const normalizedServerUrl = serverUrl.trim();

        if (!normalizedServerName || !normalizedServerUrl) {
            return;
        }

        setIsSubmitting(true);

        try {
            await addServer({
                name: normalizedServerName,
                url: normalizedServerUrl,
            });

            navigate("/dashboard");
        } finally {
            setIsSubmitting(false);
        }
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
                            disabled={isSubmitting}
                            required
                        />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="server-url">Server URL</FieldLabel>

                        <FieldDescription className="text-xs">
                            The server&apos;s URL.
                        </FieldDescription>

                        <Input
                            id="server-url"
                            type="url"
                            placeholder="https://example.com"
                            value={serverUrl}
                            onChange={(event) => setServerUrl(event.target.value)}
                            disabled={isSubmitting}
                            required
                        />
                    </Field>

                    <Button size="lg" type="submit" disabled={isSubmitting}>
                        {isSubmitting ? "Adding..." : "Add server"}
                    </Button>
                </FieldGroup>
            </FieldSet>
        </form>
    );
}

export default AddServerScreen;