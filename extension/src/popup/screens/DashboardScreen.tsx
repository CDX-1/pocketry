import { useNavigate } from "react-router-dom";
import { useAuth } from "../../components/context/auth-provider";
import { useEffect } from "react";
import { ExpiryTimer } from "../../components/expiry-timer";
import { Tooltip, TooltipContent, TooltipTrigger } from "../../components/ui/tooltip";

function DashboardScreen() {
    const { session } = useAuth();
    const navigate = useNavigate();

    useEffect(() => {
        if (!session) {
            navigate("/login");
        }
    }, [session, navigate]);

    if (!session) return null;
    const expiryDate = new Date(session.expiresAt);

    return (
        <main className="flex h-full flex-col pb-8">
            <div className="grid w-full gap-4 text-left">
                <p>
                    {session?.accessToken}
                    <br />
                    {session?.expiresAt}
                </p>
            </div>

            <div className="mt-auto pt-3 flex justify-end">
                <Tooltip>
                    <TooltipTrigger asChild>
                        <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
                            <ExpiryTimer expiresAt={expiryDate} />
                        </div>
                    </TooltipTrigger>

                    <TooltipContent side="left">
                        <p>Time until session expiry</p>
                    </TooltipContent>
                </Tooltip>
            </div>
        </main>
    );
}

export default DashboardScreen;