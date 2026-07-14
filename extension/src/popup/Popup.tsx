import { HashRouter, Route, Routes } from "react-router-dom";
import Header from "../components/header";
import LaunchScreen from "./screens/LaunchScreen";
import AddServerScreen from "./screens/AddServerScreen";
import DashboardScreen from "./screens/DashboardScreen";
import Footer from "../components/footer";
import ServerLoginScreen from "./screens/server/ServerLoginScreen";

function Popup() {
    return (
        <main className="h-[560px] w-[360px] overflow-hidden border border-border/5 bg-background text-foreground">
            <section className="flex h-full flex-col">
                <HashRouter>
                    <div className="flex-1 p-3">
                        <Header />
                        <Routes>
                            <Route path="/" element={<LaunchScreen />} />
                            <Route path="/login" element={<ServerLoginScreen />} />
                            <Route path="/add-server" element={<AddServerScreen />} />
                            <Route path="/dashboard" element={<DashboardScreen />} />
                        </Routes>
                    </div>

                    <Footer />
                </HashRouter>
            </section>
        </main>
    );
}

export default Popup;