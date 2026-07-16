import { AlertCircleIcon, InfoIcon } from "lucide-react";
import { useNotifications } from "./context/notification-provider";
import { Alert, AlertAction, AlertDescription, AlertTitle } from "./ui/alert";
import { Button } from "./ui/button";

export function NotificationViewport() {
    const { notifications, dismiss } = useNotifications();
    if (notifications.length === 0) return null;

    return (
        <div className="pointer-events-none absolute inset-x-3 top-3 z-50 space-y-2">
            {notifications.map((notif) => (
                <Alert
                    key={notif.id}
                    variant={notif.variant}
                    className="pointer-events-auto shadow-md"
                >
                    {notif.variant === "destructive"
                        ? <AlertCircleIcon />
                        : <InfoIcon />
                    }

                    <AlertTitle>{notif.title}</AlertTitle>
                    {notif.description && (
                        <AlertDescription>
                            {notif.description}
                        </AlertDescription>
                    )}

                    <AlertAction>
                        <Button
                            type="button"
                            size="xs"
                            variant={
                                notif.variant === "destructive"
                                    ? "destructive"
                                    : "outline"
                            }
                            onClick={() => dismiss(notif.id)}
                        >
                            Close
                        </Button>
                    </AlertAction>
                </Alert>
            ))}
        </div>
    );
}