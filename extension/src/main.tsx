/// <reference types="chrome" />

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import Popup from "./popup/Popup";
import { ServersProvider } from "./components/context/servers-provider";
import { AuthProvider } from "./components/context/auth-provider";
import { NotificationProvider } from "./components/context/notification-provider";
import { TooltipProvider } from "./components/ui/tooltip";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <NotificationProvider>
            <ServersProvider>
                <AuthProvider>
                    <TooltipProvider>
                        <Popup />
                    </TooltipProvider>
                </AuthProvider>
            </ServersProvider>
        </NotificationProvider>
    </StrictMode>,
);