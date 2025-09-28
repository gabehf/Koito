import { createContext, useEffect, useState, useCallback, type ReactNode } from 'react';
import { type Theme, themes } from '~/styles/themes.css';
import { themeVars } from '~/styles/vars.css';
import { useAppContext } from './AppProvider';

interface ThemeContextValue {
    themeName: string;
    theme: Theme;
    setTheme: (theme: string) => void;
    resetTheme: () => void;
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
    children,
}: {
    children: ReactNode;
}) {
<<<<<<< HEAD
    let defaultTheme = useAppContext().defaultTheme
    let initialTheme = localStorage.getItem("theme") ?? defaultTheme
=======
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
    const [themeName, setThemeName] = useState(initialTheme);
    const [currentTheme, setCurrentTheme] = useState<Theme>(() => {
        if (initialTheme === 'custom') {
            const customTheme = getStoredCustomTheme();
<<<<<<< HEAD
            return customTheme || themes[defaultTheme];
        }
        return themes[initialTheme] || themes[defaultTheme];
=======
            return customTheme || themes.yuu;
        }
        return themes[initialTheme] || themes.yuu;
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
    });

    const setTheme = (newThemeName: string) => {
        setThemeName(newThemeName);
        if (newThemeName === 'custom') {
            const customTheme = getStoredCustomTheme();
            if (customTheme) {
                setCurrentTheme(customTheme);
            } else {
                // Fallback to default theme if no custom theme found
<<<<<<< HEAD
                setThemeName(defaultTheme);
                setCurrentTheme(themes[defaultTheme]);
=======
                setThemeName('yuu');
                setCurrentTheme(themes.yuu);
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
            }
        } else {
            const foundTheme = themes[newThemeName];
            if (foundTheme) {
<<<<<<< HEAD
                localStorage.setItem('theme', newThemeName)
                setCurrentTheme(foundTheme);
            }
        }
    }

    const resetTheme = () => {
        setThemeName(defaultTheme)
        localStorage.removeItem('theme')
        setCurrentTheme(themes[defaultTheme])
=======
                setCurrentTheme(foundTheme);
            }
        }
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
    }

    const setCustomTheme = useCallback((customTheme: Theme) => {
        localStorage.setItem('custom-theme', JSON.stringify(customTheme));
        applyCustomThemeVars(customTheme);
        setThemeName('custom');
<<<<<<< HEAD
        localStorage.setItem('theme', 'custom')
=======
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
        setCurrentTheme(customTheme);
    }, []);

    const getCustomTheme = (): Theme | undefined => {
        return getStoredCustomTheme();
    }

    useEffect(() => {
        const root = document.documentElement;

        root.setAttribute('data-theme', themeName);
<<<<<<< HEAD
=======
        localStorage.setItem('theme', themeName);
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)

        if (themeName === 'custom') {
            applyCustomThemeVars(currentTheme);
        } else {
            clearCustomThemeVars();
        }
    }, [themeName, currentTheme]);

    return (
<<<<<<< HEAD
        <ThemeContext.Provider value={{ 
            themeName, 
            theme: currentTheme, 
            setTheme,
            resetTheme, 
            setCustomTheme, 
            getCustomTheme
        }}>
=======
        <ThemeContext.Provider value={{ themeName, theme: currentTheme, setTheme, setCustomTheme, getCustomTheme }}>
>>>>>>> fbaa6a6 (Rework theme provider to provide the actual Theme object throughtout the app, in addition to the name)
            {children}
        </ThemeContext.Provider>
    );
}

export { ThemeContext };
