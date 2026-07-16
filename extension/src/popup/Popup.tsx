import { HashRouter, Route, Routes } from "react-router-dom";
import Header from "../components/header";
import LaunchScreen from "./screens/LaunchScreen";
import AddServerScreen from "./screens/AddServerScreen";
import DashboardScreen from "./screens/DashboardScreen";
import Footer from "../components/footer";
import ServerLoginScreen from "./screens/server/ServerLoginScreen";
import ServerRegisterScreen from "./screens/server/ServerRegisterScreen";
import RequireAuth from "../components/context/require-auth";
import { NotificationViewport } from "../components/notification-viewport";

function Popup() {
    return (
        <main className="h-[560px] w-[360px] overflow-hidden border border-border/5 bg-background text-foreground">
            <NotificationViewport />

            <section className="flex h-full flex-col">
                <HashRouter>
                    <div className="flex-1 p-3">
                        <Header />
                        <Routes>
                            <Route path="/" element={<LaunchScreen />} />
                            <Route path="/login" element={<ServerLoginScreen />} />
                            <Route path="/register" element={<ServerRegisterScreen />} />
                            <Route path="/add-server" element={<AddServerScreen />} />

                            <Route element={<RequireAuth />}>
                                <Route path="/dashboard" element={<DashboardScreen />} />
                            </Route>
                        </Routes>
                    </div>
                    <Footer />
                </HashRouter>
            </section>
        </main>
    );
}

export default Popup;