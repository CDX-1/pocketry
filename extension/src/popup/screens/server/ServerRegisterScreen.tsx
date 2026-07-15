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
import { Alert, AlertAction, AlertDescription, AlertTitle } from "../../../components/ui/alert";
import { AlertCircleIcon } from "lucide-react";

function ServerRegisterScreen() {
    const { register } = useAuth();
    const navigate = useNavigate();

    const [error, setError] = useState("");
    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        setError("");
        setIsSubmitting(true);

        try {
            await register(username, password);

            setPassword("");

            navigate("/login", {
                replace: true,
            });
        } catch (error) {
            setError(error instanceof Error
                ? error.message
                : "Registration failed",
            );
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <form className="flex h-full flex-col justify-end pb-8 relative" onSubmit={handleSubmit}>
            {error && (
                <div className="pointer-events-none absolute inset-x-0 top-2 z-50">
                    <Alert
                        variant="destructive"
                        className="pointer-events-auto w-full shadow-sm"
                    >
                        <AlertCircleIcon />
                        <AlertTitle>Registration failed</AlertTitle>
                        <AlertDescription>
                            {error}
                        </AlertDescription>
                        <AlertAction>
                            <Button
                                type="button"
                                size="xs"
                                variant="destructive"
                                onClick={() => setError("")}
                            >
                                Close
                            </Button>
                        </AlertAction>
                    </Alert>
                </div>
            )}

            <FieldSet>
                <FieldLegend>Register an account</FieldLegend>

                <FieldDescription>
                    Create an account on this Pocketry server.
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
                        Already have an account?{" "}
                        <Link to="/login" className="underline text-primary-primary">
                            Log in to your account.
                        </Link>
                    </p>

                    <Button size="lg" type="submit" disabled={isSubmitting}>
                        {isSubmitting ? "Creating account..." : "Create account"}
                    </Button>
                </FieldGroup>
            </FieldSet>
        </form>
    );
}

export default ServerRegisterScreen;