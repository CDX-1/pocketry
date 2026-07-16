import { useNavigate } from "react-router-dom";
import { useAuth } from "../../components/context/auth-provider";

function DashboardScreen() {
    const { session } = useAuth();
    const navigate = useNavigate();

    if (!session) {
        navigate("/login");
        return;
    }

    return (
        <main className="flex h-full items-center justify-center px-6">
            <div className="grid w-full gap-4 text-left">
                <p>
                    {session?.accessToken}
                    <br />
                    {session?.expiresAt}
                </p>
            </div>
        </main>
    );
}

export default DashboardScreen;