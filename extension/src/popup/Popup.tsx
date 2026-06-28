import { HashRouter, Route, Routes } from "react-router-dom";
import Header from "../components/header";
import LaunchScreen from "./screens/LaunchScreen";
import AddServerScreen from "./screens/server/AddServerScreen";

function Popup() {
    return (
        <main className="h-[560px] w-[360px] overflow-hidden border border-border/5 bg-background p-3 text-foreground">
            <section className="flex h-full flex-col">
                <HashRouter>
                    <Header />

                    <div className="flex-1">
                        <Routes>
                            <Route path="/" element={<LaunchScreen />} />
                            <Route path="/add-server" element={<AddServerScreen />} />
                        </Routes>
                    </div>
                </HashRouter>
            </section>
        </main>
    );
}

export default Popup;