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
import { Link } from "react-router-dom";

function ServerLoginScreen() {
    const { login } = useAuth();

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        setIsSubmitting(true);

        void login(username, password);

        setIsSubmitting(false);
    }

    return (
        <form className="flex h-full flex-col justify-end pb-8" onSubmit={handleSubmit}>
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