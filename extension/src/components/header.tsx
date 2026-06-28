import { SettingsIcon } from "lucide-react";
import { Button } from "./ui/button";

function Header() {
    return (
        <header className="flex items-center justify-between">
            <h1 className="text-lg font-semibold">Pocketry</h1>

            <Button size="icon" variant="ghost" type="button">
                <SettingsIcon className="h-4 w-4" />
            </Button>
        </header>
    );
}

export default Header;