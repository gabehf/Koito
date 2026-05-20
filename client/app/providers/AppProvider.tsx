import { getCfg, type User } from "api/api";
import { createContext, useContext, useEffect, useState } from "react";
import pkg from "../../package.json";
import semver from "semver";

function isNewerVersion(current: string, latest: string): boolean {
  return semver.gt(latest, current);
}

interface AppContextType {
  user: User | null | undefined;
  configurableHomeActivity: boolean;
  homeItems: number;
  defaultTheme: string;
  currentVersion: string;
  updateAvailable: boolean;
  firstActivity: Date | undefined;
  setConfigurableHomeActivity: (value: boolean) => void;
  setHomeItems: (value: number) => void;
  setUsername: (value: string) => void;
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

  const setUsername = (value: string) => {
    if (!user) {
      return;
    }
    setUser({ ...user, username: value });
  };

  const currentVersion = import.meta.env.VITE_KOITO_VERSION || pkg.version;

  const [updateAvailable, setUpdateAvailable] = useState<boolean>(false);
  const [firstActivity, setFirstActivity] = useState<Date | undefined>();

  useEffect(() => {
    fetch("/apis/web/v1/user")
      .then((res) => res.json())
      .then((data) => {
        data.error ? setUser(null) : setUser(data);
      })
      .catch(() => setUser(null));

    setConfigurableHomeActivity(true);
    setHomeItems(12);

    getCfg().then((cfg) => {
      console.log(cfg);
      if (cfg.default_theme !== "") {
        setDefaultTheme(cfg.default_theme);
      } else {
        setDefaultTheme("yuu");
      }
    });

    fetch("/apis/web/v1/first-activity")
      .then((r) => r.json())
      .then((data) => {
        if (!data.error) {
          setFirstActivity(new Date(data.time));
        }
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    fetch("https://api.github.com/repos/gabehf/koito/releases/latest")
      .then((r) => r.json())
      .then((r) => {
        setUpdateAvailable(isNewerVersion(currentVersion, r.tag_name));
      })
      .catch((err) => console.log(err));
  }, []);

  // Block rendering the app until config is loaded
  if (user === undefined || defaultTheme === undefined) {
    return null;
  }

  const contextValue: AppContextType = {
    user,
    configurableHomeActivity,
    homeItems,
    defaultTheme,
    currentVersion,
    updateAvailable,
    firstActivity,
    setConfigurableHomeActivity,
    setHomeItems,
    setUsername,
  };

  return (
    <AppContext.Provider value={contextValue}>{children}</AppContext.Provider>
  );
};
