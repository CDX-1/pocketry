import { Navigate, Outlet, useLocation } from "react-router-dom";
import { useAuth } from "./auth-provider";

function RequireAuth() {
    const { status } = useAuth();
    const location = useLocation();

    if (status === "loading") {
        return (
            <div className="flex h-full items-center justify-center">
                <p className="text-sm text-muted-foreground">
                    Loading...
                </p>
            </div>
        );
    }

    if (status === "unauthenticated") {
        return (
            <Navigate
                to="/login"
                replace
                state={{ from: location.pathname }}
            />
        );
    }

    return <Outlet />
}

export default RequireAuth;