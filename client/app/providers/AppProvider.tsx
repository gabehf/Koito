import { getCfg, type User } from "api/api";
import {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
} from "react";
import pkg from "../../package.json";
import semver from "semver";
function isNewerVersion(current: string, latest: string): boolean {
  const a = semver.valid(current);
  const b = semver.valid(latest);
  return !!a && !!b && semver.gt(b, a);
}

interface AppContextType {
  adminSession: User | null | undefined;
  user: User | null | undefined; // Compatibility alias for existing admin controls.
  configurableHomeActivity: boolean;
  homeItems: number;
  defaultTheme: string;
  currentVersion: string;
  updateAvailable: boolean;
  setConfigurableHomeActivity: (value: boolean) => void;
  setHomeItems: (value: number) => void;
  setUsername: (value: string) => void;
  refreshAdminSession: () => Promise<void>;
}

const AppContext = createContext<AppContextType | undefined>(undefined);

export const useAppContext = () => {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error("useAppContext must be used within an AppProvider");
  }
  return context;
};

export const AppProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [defaultTheme, setDefaultTheme] = useState<string | undefined>(
    undefined,
  );
  const [configurableHomeActivity, setConfigurableHomeActivity] =
    useState<boolean>(false);
  const [homeItems, setHomeItems] = useState<number>(0);

  const refreshAdminSession = useCallback(async () => {
    const response = await fetch("/apis/web/v1/user");
    if (response.status === 401) {
      setUser(null);
      return;
    }
    if (!response.ok)
      throw new Error("Unable to refresh administrator session");
    setUser(await response.json());
  }, []);

  const setUsername = (value: string) => {
    if (!user) {
      return;
    }
    setUser({ ...user, username: value });
  };

  const currentVersion = import.meta.env.VITE_KOITO_VERSION || pkg.version;

  const [updateAvailable, setUpdateAvailable] = useState<boolean>(false);

  useEffect(() => {
    refreshAdminSession().catch(() => setUser(null));

    setConfigurableHomeActivity(true);
    setHomeItems(12);

    getCfg()
      .then((cfg) => {
        if (cfg.default_theme !== "") {
          setDefaultTheme(cfg.default_theme);
        } else {
          setDefaultTheme("yuu");
        }
      })
      .catch(() => setDefaultTheme("yuu"));
  }, []);

  useEffect(() => {
    fetch("https://api.github.com/repos/gabehf/Koito/releases/latest")
      .then((response) => response.json())
      .then((r) => {
        setUpdateAvailable(!!r && isNewerVersion(currentVersion, r.tag_name));
      })
      .catch((err) => console.log(err));
  }, []);

  // Block rendering the app until config is loaded
  if (user === undefined || defaultTheme === undefined) {
    return null;
  }

  const contextValue: AppContextType = {
    adminSession: user,
    user,
    configurableHomeActivity,
    homeItems,
    defaultTheme,
    currentVersion,
    updateAvailable,
    setConfigurableHomeActivity,
    setHomeItems,
    setUsername,
    refreshAdminSession,
  };

  return (
    <AppContext.Provider value={contextValue}>{children}</AppContext.Provider>
  );
};
