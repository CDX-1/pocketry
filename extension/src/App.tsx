function App() {
    return (
        <main className="min-h-screen w-[360px] bg-background p-4 text-foreground">
            <section className="space-y-4">
                <div>
                    <h1 className="text-xl font-semibold">Pocketry</h1>
                    <p className="text-sm text-muted-foreground">
                        Unlock your vault to continue.
                    </p>
                </div>

                <form className="space-y-3">
                    <div className="space-y-1">
                        <label className="text-sm font-medium" htmlFor="email">
                            Email
                        </label>
                        <input
                            id="email"
                            className="w-full rounded-md border px-3 py-2 text-sm"
                            type="email"
                            placeholder="you@example.com"
                        />
                    </div>

                    <div className="space-y-1">
                        <label className="text-sm font-medium" htmlFor="password">
                            Password
                        </label>
                        <input
                            id="password"
                            className="w-full rounded-md border px-3 py-2 text-sm"
                            type="password"
                            placeholder="Master password"
                        />
                    </div>

                    <button
                        className="w-full rounded-md bg-black px-3 py-2 text-sm font-medium text-white"
                        type="submit"
                    >
                        Unlock Vault
                    </button>
                </form>
            </section>
        </main>
    );
}

export default App;