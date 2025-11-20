import type { Theme } from "~/styles/themes.css";

interface Props {
    theme: Theme
    themeName: string
    setTheme: Function
}

export default function ThemeOption({ theme, themeName, setTheme }: Props) {

    const capitalizeFirstLetter = (s: string) => {
        return s.charAt(0).toUpperCase() + s.slice(1);
    }

    return (
        <div onClick={() => setTheme(themeName)} className="rounded-md p-3 sm:p-5 hover:cursor-pointer flex gap-4 items-center border-2" style={{background: theme.bg, color: theme.fg, borderColor: theme.bgSecondary}}>
            <div className="text-xs sm:text-sm">{capitalizeFirstLetter(themeName)}</div>
            <div className="w-[50px] h-[30px] rounded-md" style={{background: theme.bgSecondary}}></div>
            <div className="w-[50px] h-[30px] rounded-md" style={{background: theme.fgSecondary}}></div>
            <div className="w-[50px] h-[30px] rounded-md" style={{background: theme.primary}}></div>
        </div>
    )
}