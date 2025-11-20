import { getCfg, type User } from "api/api";
import { createContext, useContext, useEffect, useState } from "react";

interface AppContextType {
  user: User | null | undefined;
  configurableHomeActivity: boolean;
  homeItems: number;
  defaultTheme: string;
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
    undefined
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

  useEffect(() => {
    fetch("/apis/web/v1/user/me")
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
    setConfigurableHomeActivity,
    setHomeItems,
    setUsername,
  };

  return (
    <AppContext.Provider value={contextValue}>{children}</AppContext.Provider>
  );
};
