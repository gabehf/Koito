import type { ReactNode } from "react";
import Sidebar from "./sidebar/Sidebar";
import Footer from "./Footer";
export default function AppShell({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col sm:flex-row min-h-screen">
      <Sidebar />
      <div className="flex flex-col items-center mx-auto w-full ml-0 sm:ml-[58px] min-w-0">
        {children}
        <Footer />
      </div>
    </div>
  );
}
