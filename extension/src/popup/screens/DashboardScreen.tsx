import { type FormEvent, useState } from "react";
import { LockIcon } from "lucide-react";

import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";

function DashboardScreen() {
    const [masterPassword, setMasterPassword] = useState("");
    const [isUnlocking, setIsUnlocking] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        const password = masterPassword.trim();

        if (!password) {
            return;
        }

        setIsUnlocking(true);

        try {
            console.log("Unlock vault");
            setMasterPassword("");
        } finally {
            setIsUnlocking(false);
        }
    }

    return (
        <main className="flex h-full items-center justify-center px-6">
            <form
                onSubmit={handleSubmit}
                className="flex w-full max-w-sm flex-col items-center text-center"
            >
                <div className="mb-6 flex flex-col items-center">
                    <div className="mb-3 flex size-12 items-center justify-center rounded-full bg-muted">
                        <LockIcon className="size-5 text-muted-foreground" />
                    </div>

                    <h1 className="text-lg font-semibold tracking-tight">
                        Vault locked
                    </h1>

                    <p className="mt-1 max-w-[260px] text-sm leading-5 text-muted-foreground">
                        Enter your master password to unlock Pocketry.
                    </p>
                </div>

                <div className="grid w-full gap-4 text-left">
                    <div className="grid gap-2">
                        <Label htmlFor="master-password">
                            Master password
                        </Label>

                        <Input
                            id="master-password"
                            type="password"
                            value={masterPassword}
                            onChange={(event) =>
                                setMasterPassword(event.target.value)
                            }
                            placeholder="Enter password"
                            autoComplete="current-password"
                            disabled={isUnlocking}
                            required
                        />
                    </div>

                    <Button
                        type="submit"
                        className="w-full"
                        disabled={isUnlocking || !masterPassword.trim()}
                    >
                        {isUnlocking ? "Unlocking..." : "Unlock vault"}
                    </Button>
                </div>
            </form>
        </main>
    );
}

export default DashboardScreen;