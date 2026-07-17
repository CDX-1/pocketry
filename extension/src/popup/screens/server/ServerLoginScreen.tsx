import { useState, type FormEvent } from "react";
import { Button } from "../../../components/ui/button";
import {
    Field,
    FieldDescription,
    FieldGroup,
    FieldLabel,
    FieldLegend,
    FieldSet,
} from "../../../components/ui/field";
import { Input } from "../../../components/ui/input";
import { useAuth } from "../../../components/context/auth-provider";
import { Link, useNavigate } from "react-router-dom";
import { useNotifications } from "../../../components/context/notification-provider";

function ServerLoginScreen() {
    const { login } = useAuth();
    const { notify } = useNotifications();
    const navigate = useNavigate();

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        setIsSubmitting(true);

        try {
            await login(username, password);

            setPassword("");

            navigate("/dashboard", {
                replace: true,
            });
        } catch (error) {
            notify({
                title: "Login failed",
                description: error instanceof Error ? error.message : "Login failed",
                variant: "destructive",
                ttl: 0,
            });
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <form className="flex h-full flex-col justify-end pb-8 relative" onSubmit={handleSubmit}>
            <FieldSet>
                <FieldLegend>Log in to your account</FieldLegend>

                <FieldDescription>
                    Log in to this Pocketry server.
                </FieldDescription>

                <FieldGroup>
                    <Field>
                        <FieldLabel htmlFor="username">Username</FieldLabel>

                        <FieldDescription className="text-xs">
                            Account username.
                        </FieldDescription>

                        <Input
                            id="username"
                            type="text"
                            value={username}
                            onChange={(event) => setUsername(event.target.value)}
                            disabled={isSubmitting}
                            required
                        />
                    </Field>

                    <Field>
                        <FieldLabel htmlFor="password">Password</FieldLabel>

                        <FieldDescription className="text-xs">
                            Account password.
                        </FieldDescription>

                        <Input
                            id="password"
                            type="password"
                            value={password}
                            onChange={(event) => setPassword(event.target.value)}
                            disabled={isSubmitting}
                            required
                        />
                    </Field>

                    <p className="text-center text-sm">
                        Don't have an account?{" "}
                        <Link to="/register" className="underline text-primary-primary">
                            Register an account.
                        </Link>
                    </p>

                    <Button size="lg" type="submit" disabled={isSubmitting}>
                        {isSubmitting ? "Logging in..." : "Login"}
                    </Button>
                </FieldGroup>
            </FieldSet>
        </form>
    );
}

export default ServerLoginScreen;