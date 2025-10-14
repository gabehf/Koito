import { createContext, useEffect, useState, useCallback, type ReactNode } from 'react';
import { type Theme, themes } from '~/styles/themes.css';
import { themeVars } from '~/styles/vars.css';

interface ThemeContextValue {
    themeName: string;
    theme: Theme;
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

function getStoredCustomTheme(): Theme | undefined {
    const themeStr = localStorage.getItem('custom-theme');
    if (!themeStr) return undefined;
    try {
        const parsed = JSON.parse(themeStr);
        const { name, ...theme } = parsed;
        return theme as Theme;
    } catch {
        return undefined;
    }
}

export function ThemeProvider({
    theme: initialTheme,
    children,
}: {
    theme: string;
    children: ReactNode;
}) {
    const [themeName, setThemeName] = useState(initialTheme);
    const [currentTheme, setCurrentTheme] = useState<Theme>(() => {
        if (initialTheme === 'custom') {
            const customTheme = getStoredCustomTheme();
            return customTheme || themes.yuu;
        }
        return themes[initialTheme] || themes.yuu;
    });

    const setTheme = (newThemeName: string) => {
        setThemeName(newThemeName);
        if (newThemeName === 'custom') {
            const customTheme = getStoredCustomTheme();
            if (customTheme) {
                setCurrentTheme(customTheme);
            } else {
                // Fallback to default theme if no custom theme found
                setThemeName('yuu');
                setCurrentTheme(themes.yuu);
            }
        } else {
            const foundTheme = themes[newThemeName];
            if (foundTheme) {
                setCurrentTheme(foundTheme);
            }
        }
    }

    const setCustomTheme = useCallback((customTheme: Theme) => {
        localStorage.setItem('custom-theme', JSON.stringify(customTheme));
        applyCustomThemeVars(customTheme);
        setThemeName('custom');
        setCurrentTheme(customTheme);
    }, []);

    const getCustomTheme = (): Theme | undefined => {
        return getStoredCustomTheme();
    }

    useEffect(() => {
        const root = document.documentElement;

        root.setAttribute('data-theme', themeName);
        localStorage.setItem('theme', themeName);

        if (themeName === 'custom') {
            applyCustomThemeVars(currentTheme);
        } else {
            clearCustomThemeVars();
        }
    }, [themeName, currentTheme]);

    return (
        <ThemeContext.Provider value={{ themeName, theme: currentTheme, setTheme, setCustomTheme, getCustomTheme }}>
            {children}
        </ThemeContext.Provider>
    );
}

export { ThemeContext };
