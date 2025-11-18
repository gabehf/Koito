import {
  isRouteErrorResponse,
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useRouteError,
} from "react-router";

import type { Route } from "./+types/root";
import './themes.css'
import "./app.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ThemeProvider } from './providers/ThemeProvider';
import Sidebar from "./components/sidebar/Sidebar";
import Footer from "./components/Footer";
import { AppProvider } from "./providers/AppProvider";

// Create a client
const queryClient = new QueryClient()

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:ital,opsz,wght@0,14..32,100..900;1,14..32,100..900&display=swap",
  },
];

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" style={{backgroundColor: 'black'}}>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <link rel="icon" type="image/png" href="/favicon-96x96.png" sizes="96x96" />
        <link rel="icon" type="image/svg+xml" href="/favicon.svg" />
        <link rel="shortcut icon" href="/favicon.ico" />
        <link rel="apple-touch-icon" sizes="180x180" href="/apple-touch-icon.png" />
        <meta name="apple-mobile-web-app-title" content="Koito" />
        <link rel="manifest" href="/site.webmanifest" />
        <Meta />
        <Links />
      </head>
      <body className="min-h-screen">
        {children}
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}

export default function App() {
  return (
    <>
        <AppProvider>
            <ThemeProvider>
            <QueryClientProvider client={queryClient}>
                <div className="flex-col flex sm:flex-row">
                  <Sidebar />
                  <div className="flex flex-col items-center mx-auto w-full ml-0 sm:ml-[40px]">
                      <Outlet />
                      <Footer />
                  </div>
                </div>
            </QueryClientProvider>
            </ThemeProvider>
        </AppProvider>
    </>
  );
}

export function HydrateFallback() {
    return null
}

export function ErrorBoundary() {
    const error = useRouteError();
    let message = "Oops!";
    let details = "An unexpected error occurred.";
    let stack: string | undefined;

    if (isRouteErrorResponse(error)) {
        message = error.status === 404 ? "404" : "Error";
        details = error.status === 404
        ? "The requested page could not be found."
        : error.statusText || details;
    } else if (import.meta.env.DEV && error instanceof Error) {
        details = error.message;
        stack = error.stack;
    }


    const title = `${message} - Koito`

    return (
        <AppProvider>
            <ThemeProvider>
            <title>{title}</title>
                <div className="flex">
                    <Sidebar />
                    <div className="w-full flex flex-col">
                        <main className="pt-16 p-4 container mx-auto flex-grow">
                            <div className="flex gap-4 items-end">
                                <img className="w-[200px] rounded" src="../yuu.jpg" />
                                <div>
                                    <h1>{message}</h1>
                                    <p>{details}</p>
                                </div>
                            </div>
                            {stack && (
                                <pre className="w-full p-4 overflow-x-auto">
                                <code>{stack}</code>
                                </pre>
                            )}
                        </main>
                        <Footer />
                    </div>
                </div>
            </ThemeProvider>
        </AppProvider>
    );
}
