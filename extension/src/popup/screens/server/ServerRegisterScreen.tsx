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

function ServerRegisterScreen() {
    const { register } = useAuth();
    const { notify } = useNotifications();
    const navigate = useNavigate();

    const [username, setUsername] = useState("");
    const [password, setPassword] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    async function handleSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();

        setIsSubmitting(true);

        try {
            await register(username, password);

            setPassword("");

            notify({
                title: "Registration success",
                description: "Account created successfully. Please log in.",
                variant: "default",
            });

            navigate("/login", {
                replace: true,
            });
        } catch (error) {
            notify({
                title: "Registration failed",
                description: error instanceof Error ? error.message : "Registration failed",
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