import { useOptionalProfile } from "~/providers/ProfileProvider";
import { Modal } from "./Modal";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs";
import AccountPage from "./AccountPage";
import { ThemeSwitcher } from "../themeSwitcher/ThemeSwitcher";
import AdministrationPanel from "../AdministrationPanel";
import { useRef } from "react";
import { useRevalidator } from "react-router";
import { useQueryClient } from "@tanstack/react-query";
import { useAppContext } from "~/providers/AppProvider";
import ApiKeysModal from "./ApiKeysModal";
import ExportModal from "./ExportModal";
import About from "./About";

interface Props {
  open: boolean;
  setOpen: Function;
}

export default function SettingsModal({ open, setOpen }: Props) {
  const { user, updateAvailable } = useAppContext();
  const activeProfile = useOptionalProfile();
  const queryClient = useQueryClient();
  const revalidator = useRevalidator();
  const adminVisited = useRef(false);
  const close = () => {
    setOpen(false);
    if (adminVisited.current) {
      adminVisited.current = false;
      // Administrative edits may affect the public page behind this dialog.
      if (activeProfile)
        void queryClient.invalidateQueries({
          queryKey: ["profile", activeProfile.profile.id],
        });
      void revalidator.revalidate();
    }
  };

  const triggerClasses =
    "px-1 sm:px-4 py-2 w-full hover-bg-secondary " +
    "rounded-md text-start border-(--color-border) " +
    "data-[state=active]:bg-[var(--color-bg-secondary)] data-[state=active]:border";
  const contentClasses =
    "w-full min-w-0 px-2 mt-8 sm:mt-4 sm:px-10 overflow-y-auto sm:mx-3";

  return (
    <Modal label="Settings" h={700} isOpen={open} onClose={close} maxW={1000}>
      <Tabs
        defaultValue="Appearance"
        onValueChange={(value) => {
          if (value === "Administration") adminVisited.current = true;
        }}
        orientation="vertical"
        className="flex flex-col sm:flex-row h-full"
      >
        <div className="flex flex-col gap-1 mx-auto w-17/20 sm:max-w-1/4 rounded-md bg p-2 border">
          <TabsList
            aria-label="Settings sections"
            className="flex flex-col gap-1"
          >
            <TabsTrigger className={triggerClasses} value="Appearance">
              Appearance
            </TabsTrigger>
            <TabsTrigger className={triggerClasses} value="Account">
              Account
            </TabsTrigger>
            {user && activeProfile && (
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
            <TabsTrigger className={triggerClasses} value="Administration">
              Administration
            </TabsTrigger>
          </TabsList>
        </div>

        <TabsContent value="Administration" className={contentClasses}>
          <AdministrationPanel
            embedded
            onNavigate={() => {
              adminVisited.current = false;
              setOpen(false);
            }}
          />
        </TabsContent>
        <TabsContent value="Account" className={contentClasses}>
          <AccountPage />
        </TabsContent>
        <TabsContent value="Appearance" className={contentClasses}>
          <ThemeSwitcher />
        </TabsContent>
        <TabsContent value="API Keys" className={contentClasses}>
          {user && activeProfile && <ApiKeysModal />}
        </TabsContent>
        <TabsContent value="Export" className={contentClasses}>
          {user && activeProfile && <ExportModal />}
        </TabsContent>
        <TabsContent value="About" className={contentClasses}>
          <About />
        </TabsContent>
      </Tabs>
    </Modal>
  );
}
