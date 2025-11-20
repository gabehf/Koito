import { useState } from "react";
import { useTheme } from "../../hooks/useTheme";
import themes from "~/styles/themes.css";
import ThemeOption from "./ThemeOption";
import { AsyncButton } from "../AsyncButton";

export function ThemeSwitcher() {
  const { setTheme } = useTheme();
  const initialTheme = {
    bg: "#1e1816",
    bgSecondary: "#2f2623",
    bgTertiary: "#453733",
    fg: "#f8f3ec",
    fgSecondary: "#d6ccc2",
    fgTertiary: "#b4a89c",
    primary: "#f5a97f",
    primaryDim: "#d88b65",
    accent: "#f9db6d",
    accentDim: "#d9bc55",
    error: "#e26c6a",
    warning: "#f5b851",
    success: "#8fc48f",
    info: "#87b8dd",
  };

  const { setCustomTheme, getCustomTheme, resetTheme } = useTheme();
  const [custom, setCustom] = useState(
    JSON.stringify(getCustomTheme() ?? initialTheme, null, "  ")
  );

  const handleCustomTheme = () => {
    console.log(custom);
    try {
      const themeData = JSON.parse(custom);
      setCustomTheme(themeData);
      setCustom(JSON.stringify(themeData, null, "  "));
      console.log(themeData);
    } catch (err) {
      console.log(err);
    }
  };

  return (
    <div className="flex flex-col gap-10">
      <div>
        <div className="flex items-center gap-3">
          <h2>Select Theme</h2>
          <div className="mb-3">
            <AsyncButton onClick={resetTheme}>Reset</AsyncButton>
          </div>
        </div>
        <div className="grid grid-cols-2 items-center gap-2">
          {Object.entries(themes).map(([name, themeData]) => (
            <ThemeOption
              setTheme={setTheme}
              key={name}
              theme={themeData}
              themeName={name}
            />
          ))}
        </div>
      </div>
      <div>
        <h2>Use Custom Theme</h2>
        <div className="flex flex-col items-center gap-3 bg-secondary p-5 rounded-lg">
          <textarea
            name="custom-theme"
            onChange={(e) => setCustom(e.target.value)}
            id="custom-theme-input"
            className="bg-(--color-bg) h-[450px] w-[300px] p-5 rounded-md"
            value={custom}
          />
          <AsyncButton onClick={handleCustomTheme}>Submit</AsyncButton>
        </div>
      </div>
    </div>
  );
}
