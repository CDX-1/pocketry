import { createContext, useCallback, useContext, useEffect, useMemo, useRef, useState, type ReactNode } from "react";

type NotificationVariant = "default" | "destructive";

type Notification = {
    id: string;
    title: string;
    description?: string;
    variant?: NotificationVariant;
    ttl?: number;
};

type ShowNotificationInput = Omit<Notification, "id">;

type NotificationContextValue = {
    notifications: Notification[];
    notify: (notification: ShowNotificationInput) => string;
    dismiss: (id: string) => void;
    clearNotifications: () => void;
}

const NotificationContext = createContext<NotificationContextValue | null>(null);

export function NotificationProvider({ children }: { children: ReactNode }) {
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const timersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

    const clearTimer = useCallback((id: string) => {
        const timer = timersRef.current.get(id);

        if (timer !== undefined) {
            clearTimeout(timer);
            timersRef.current.delete(id);
        }
    }, []);

    const dismiss = useCallback((id: string) => {
        clearTimer(id);
        setNotifications((prev) => prev.filter((notif) => notif.id !== id));
    }, [clearTimer]);

    const notify = useCallback((input: ShowNotificationInput): string => {
        const id = crypto.randomUUID();

        const notif: Notification = {
            id,
            variant: "default",
            ttl: 5000,
            ...input,
        };

        setNotifications((prev) => [
            ...prev,
            notif,
        ]);

        if (notif.ttl !== undefined && notif.ttl > 0) {
            const timer = setTimeout(() => {
                timersRef.current.delete(id);

                setNotifications((prev) => prev.filter(
                    (item) => item.id !== id,
                ));
            }, notif.ttl);

            timersRef.current.set(id, timer);
        }

        return id;
    }, []);

    const clearNotifications = useCallback(() => {
        for (const timer of timersRef.current.values()) {
            clearTimeout(timer);
        }

        timersRef.current.clear();
        setNotifications([]);
    }, []);

    useEffect(() => {
        return () => {
            for (const timer of timersRef.current.values()) {
                clearTimeout(timer);
            }

            timersRef.current.clear();
        };
    }, []);

    const value = useMemo(
        () => ({
            notifications,
            notify,
            dismiss,
            clearNotifications,
        }),
        [
            notifications,
            notify,
            dismiss,
            clearNotifications,
        ]
    );

    return (
        <NotificationContext.Provider value={value}>
            {children}
        </NotificationContext.Provider>
    );
}

export function useNotifications() {
    const context = useContext(NotificationContext);

    if (!context) {
        throw new Error("useNotifications must be used within a NotificationProvider");
    }

    return context;
}