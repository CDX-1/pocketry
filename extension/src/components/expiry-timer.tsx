import { Clock } from "lucide-react";
import { useEffect, useState } from "react";

const formatTime = (ms: number) => {
    if (ms <= 0) return "Expired";

    const totalSecs = Math.floor(ms / 1000);
    const hours = Math.floor(totalSecs / 3600);
    const minutes = Math.floor((totalSecs % 3600) / 60);
    const seconds = totalSecs % 60;
    
    const pad = (num: number) => String(num).padStart(2, "0");

    return hours > 0
        ? `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`
        : `${pad(minutes)}:${pad(seconds)}`;
};

export function ExpiryTimer({ expiresAt }: { expiresAt: Date }) {
    const [timeLeft, setTimeLeft] = useState(() =>
        Math.max(0, expiresAt.getTime() - Date.now()),
    );

    useEffect(() => {
        if (timeLeft <= 0) return;

        const intervalId = setInterval(() => {
            const remaining = expiresAt.getTime() - Date.now();

            if (remaining <= 0) {
                setTimeLeft(0);
                clearInterval(intervalId);
            } else {
                setTimeLeft(remaining);
            }
        }, 1000);

        return () => clearTimeout(intervalId);
    }, [expiresAt]);

    return (
        <div className="flex items-center gap-2 font-mono text-sm">
            <Clock className="w-4 h-4" />
            <span>{formatTime(timeLeft)}</span>
        </div>
    );
} 