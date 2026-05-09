import { Modal } from "./Modal";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs";
import AccountPage from "./AccountPage";
import { ThemeSwitcher } from "../themeSwitcher/ThemeSwitcher";
import ThemeHelper from "../../routes/ThemeHelper";
import { useAppContext } from "~/providers/AppProvider";
import ApiKeysModal from "./ApiKeysModal";
import { AsyncButton } from "../AsyncButton";
import ExportModal from "./ExportModal";
import About from "./About";

interface Props {
  open: boolean;
  setOpen: Function;
}

export default function SettingsModal({ open, setOpen }: Props) {
  const { user, updateAvailable } = useAppContext();

  const triggerClasses =
    "px-1 sm:px-4 py-2 w-full hover-bg-secondary " +
    "rounded-md text-start border-(--color-border) " +
    "data-[state=active]:bg-[var(--color-bg-secondary)] data-[state=active]:border";
  const contentClasses =
    "w-full px-2 mt-8 sm:mt-4 sm:px-10 overflow-y-auto sm:mx-3";

  return (
    <Modal h={700} isOpen={open} onClose={() => setOpen(false)} maxW={1000}>
      <Tabs
        defaultValue="Appearance"
        orientation="vertical"
        className="flex flex-col sm:flex-row h-full"
      >
        <TabsList className="flex flex-col gap-1 mx-auto w-17/20 sm:max-w-1/4 rounded-md bg p-2 border">
          <TabsTrigger className={triggerClasses} value="Appearance">
            Appearance
          </TabsTrigger>
          <TabsTrigger className={triggerClasses} value="Account">
            Account
          </TabsTrigger>
          {user && (
            <>
              <TabsTrigger className={triggerClasses} value="API Keys">
                API Keys
              </TabsTrigger>
              <TabsTrigger className={triggerClasses} value="Export">
                Export
              </TabsTrigger>
            </>
          )}
          <TabsTrigger
            className={`${triggerClasses} inline-flex items-center gap-2`}
            value="About"
          >
            About{" "}
            {updateAvailable && (
              <div className="h-1.5 w-1.5 rounded-full bg-(--color-info)" />
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="Account" className={contentClasses}>
          <AccountPage />
        </TabsContent>
        <TabsContent value="Appearance" className={contentClasses}>
          <ThemeSwitcher />
        </TabsContent>
        <TabsContent value="API Keys" className={contentClasses}>
          <ApiKeysModal />
        </TabsContent>
        <TabsContent value="Export" className={contentClasses}>
          <ExportModal />
        </TabsContent>
        <TabsContent value="About" className={contentClasses}>
          <About />
        </TabsContent>
      </Tabs>
    </Modal>
  );
}
