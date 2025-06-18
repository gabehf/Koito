import { createContext, useEffect, useState, useCallback, type ReactNode } from 'react';
import { type Theme } from '~/styles/themes.css';
import { themeVars } from '~/styles/vars.css';

interface ThemeContextValue {
    theme: string;
    setTheme: (theme: string) => void;
    setCustomTheme: (theme: Theme) => void;
    getCustomTheme: () => Theme | undefined;
}

const ThemeContext = createContext<ThemeContextValue | undefined>(undefined);

function toKebabCase(str: string) {
    return str.replace(/[A-Z]/g, m => '-' + m.toLowerCase());
}

function applyCustomThemeVars(theme: Theme) {
    const root = document.documentElement;
    for (const [key, value] of Object.entries(theme)) {
        if (key === 'name') continue;
        root.style.setProperty(`--color-${toKebabCase(key)}`, value);
    }
}

function clearCustomThemeVars() {
    for (const cssVar of Object.values(themeVars)) {
        document.documentElement.style.removeProperty(cssVar);
    }
}

export function ThemeProvider({
    theme: initialTheme,
    children,
}: {
    theme: string;
    children: ReactNode;
}) {
    const [theme, setThemeName] = useState(initialTheme);

    const setTheme = (theme: string) => {
      setThemeName(theme)
    }

    const setCustomTheme = useCallback((customTheme: Theme) => {
        localStorage.setItem('custom-theme', JSON.stringify(customTheme));
        applyCustomThemeVars(customTheme);
        setTheme('custom');
    }, []);

    const getCustomTheme = (): Theme | undefined => { 
        const themeStr = localStorage.getItem('custom-theme');
        if (!themeStr) {
            return undefined
        }
        try {
            let theme = JSON.parse(themeStr) as Theme
            return theme
        } catch (err) {
            return undefined 
        }
    }

    useEffect(() => {
        const root = document.documentElement;

        root.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme)
        console.log(theme)

        if (theme === 'custom') {
            const saved = localStorage.getItem('custom-theme');
            if (saved) {
                try {
                    const parsed = JSON.parse(saved) as Theme;
                    applyCustomThemeVars(parsed);
                } catch (err) {
                    console.error('Invalid custom theme in localStorage', err);
                }
            } else {
                setTheme('yuu')
            }
        } else {
            clearCustomThemeVars()
        }
    }, [theme]);

    return (
        <ThemeContext.Provider value={{ theme, setTheme, setCustomTheme, getCustomTheme }}>
            {children}
        </ThemeContext.Provider>
    );
}

export { ThemeContext };
