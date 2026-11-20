import type { ReactNode } from "react";
import {
    Links,
    Meta,
    Outlet,
    Scripts,
    ScrollRestoration,
    isRouteErrorResponse,
} from "react-router";
import type { Route } from "./+types/root";
import "./index.scss";

export function Layout({ children }: { children: ReactNode }) {
    return (
        <html lang="en">
            <head>
                <meta charSet="UTF-8" />
                <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
                <meta
                    name="viewport"
                    content="width=device-width, initial-scale=1.0"
                />
                <title>Waystore</title>
                <Meta />
                <Links />
            </head>
            <body>
                {children}
                <ScrollRestoration />
                <Scripts />
            </body>
        </html>
    );
}

export default function App() {
    return <Outlet />;
}

export function HydrateFallback() {
    return <p>Loading…</p>;
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
    if (isRouteErrorResponse(error)) {
        return (
            <main>
                <h1>
                    {error.status === 404
                        ? "Page not found"
                        : `Error ${error.status}`}
                </h1>
                <p>{error.statusText}</p>
            </main>
        );
    }

    return (
        <main>
            <h1>Something went wrong</h1>
            {import.meta.env.DEV && error instanceof Error && (
                <pre>{error.stack}</pre>
            )}
        </main>
    );
}
