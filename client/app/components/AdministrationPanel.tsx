import AdminProfiles from "./AdminProfiles";
import { useAppContext } from "~/providers/AppProvider";
import AccountPage from "./modals/AccountPage";

export default function AdministrationPanel({
  embedded = false,
  onNavigate,
}: {
  embedded?: boolean;
  onNavigate?: () => void;
}) {
  const { adminSession } = useAppContext();
  const Heading = embedded ? "h2" : "h1";
  return (
    <section className="min-w-0">
      <Heading className="text-2xl">Administration</Heading>
      <p className="my-5 text-(--color-fg-secondary)">
        The administrator account is separate from public listening profiles.
      </p>
      {adminSession ? (
        <>
          <AdminProfiles embedded={embedded} onNavigate={onNavigate} />
          {!embedded && (
            <details className="border rounded-lg p-4 mt-10">
              <summary className="cursor-pointer">
                Administrator account ({adminSession.username})
              </summary>
              <div className="mt-6">
                <AccountPage />
              </div>
            </details>
          )}
          {embedded && (
            <p className="mt-6 text-sm">
              Use the Account tab to change the administrator login or password.
            </p>
          )}
        </>
      ) : (
        <AccountPage />
      )}
    </section>
  );
}
