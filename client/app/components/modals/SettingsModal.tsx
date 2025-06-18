import { Modal } from "./Modal"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@radix-ui/react-tabs";
import AccountPage from "./AccountPage";
import { ThemeSwitcher } from "../themeSwitcher/ThemeSwitcher";
import ThemeHelper from "../../routes/ThemeHelper";
import { useAppContext } from "~/providers/AppProvider";
import ApiKeysModal from "./ApiKeysModal";
import { AsyncButton } from "../AsyncButton";
import ExportModal from "./ExportModal";

interface Props {
    open: boolean 
    setOpen: Function
}

export default function SettingsModal({ open, setOpen } : Props) {

    const { user } = useAppContext()

    const triggerClasses = "px-4 py-2 w-full hover-bg-secondary rounded-md text-start data-[state=active]:bg-[var(--color-bg-secondary)]"
    const contentClasses = "w-full px-2 mt-8 sm:mt-0 sm:px-10 overflow-y-auto"

    return (
        <Modal h={700} isOpen={open} onClose={() => setOpen(false)} maxW={900}>
            <Tabs
                defaultValue="Appearance"
                orientation="vertical" // still vertical, but layout is responsive via Tailwind
                className="flex flex-col sm:flex-row h-full"
            >
                <TabsList className="flex flex-row sm:flex-col gap-1 w-full sm:max-w-1/4 rounded-md bg p-2">
                    <TabsTrigger className={triggerClasses} value="Appearance">Appearance</TabsTrigger>
                    <TabsTrigger className={triggerClasses} value="Account">Account</TabsTrigger>
                    {user && (
                        <>
                            <TabsTrigger className={triggerClasses} value="API Keys">
                                API Keys
                            </TabsTrigger>
                            <TabsTrigger className={triggerClasses} value="Export">Export</TabsTrigger>
                        </>
                    )}
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
            </Tabs>
        </Modal>
    )
}