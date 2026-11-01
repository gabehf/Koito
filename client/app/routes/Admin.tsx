import { Link } from "react-router";
import AdministrationPanel from "~/components/AdministrationPanel";
import AppShell from "~/components/AppShell";

export default function Admin() {
  return (
    <AppShell>
      <main className="max-w-2xl w-full px-6 py-12 mx-auto min-h-[80vh]">
        <AdministrationPanel />
        <Link to="/" className="inline-block underline mt-8">
          Choose a listening profile
        </Link>
      </main>
    </AppShell>
  );
}
