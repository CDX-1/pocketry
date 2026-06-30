/// <reference types="chrome" />

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import Popup from "./popup/Popup";
import { ServersProvider } from "./components/context/servers-provider";

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <ServersProvider>
            <Popup />
        </ServersProvider>
    </StrictMode>,
);