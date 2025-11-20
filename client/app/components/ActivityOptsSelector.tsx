import { ChevronDown, ChevronUp } from "lucide-react";
import { useEffect, useState } from "react";

interface Props {
    stepSetter: (value: string) => void;
    currentStep: string;
    rangeSetter: (value: number) => void;
    currentRange: number;
    disableCache?: boolean;
}

export default function ActivityOptsSelector({
    stepSetter,
    currentStep,
    rangeSetter,
    currentRange,
    disableCache = false,
}: Props) {
    const stepPeriods = ['day', 'week', 'month'];
    const rangePeriods = [105, 182, 364];
    const [collapsed, setCollapsed] = useState(true);

    const setMenuOpen = (val: boolean) => {
        setCollapsed(val)
        if (!disableCache) {
            localStorage.setItem('activity_configuring_' + window.location.pathname.split('/')[1], String(!val));
        }
    }

    const setStep = (val: string) => {
        stepSetter(val);
        if (!disableCache) {
            localStorage.setItem('activity_step_' + window.location.pathname.split('/')[1], val);
        }
    };

    const setRange = (val: number) => {
        rangeSetter(val);
        if (!disableCache) {
            localStorage.setItem('activity_range_' + window.location.pathname.split('/')[1], String(val));
        }
    };

    useEffect(() => {
        if (!disableCache) {
            // TODO: the '182' here overwrites the default range as configured in the ActivityGrid. This is bad. Only one of these should determine the default.
            const cachedRange = parseInt(localStorage.getItem('activity_range_' + window.location.pathname.split('/')[1]) ?? '182');
            if (cachedRange) rangeSetter(cachedRange);
            const cachedStep = localStorage.getItem('activity_step_' + window.location.pathname.split('/')[1]);
            if (cachedStep) stepSetter(cachedStep);
            const cachedConfiguring = localStorage.getItem('activity_configuring_' + window.location.pathname.split('/')[1]);
            if (cachedStep) setMenuOpen(cachedConfiguring !== "true");
        }
    }, []);

    return (
        <div className="relative w-full">
            <button
                onClick={() => setMenuOpen(!collapsed)}
                className="absolute left-[75px] -top-9 text-muted hover:color-fg transition"
                title="Toggle options"
            >
                {collapsed ? <ChevronDown size={18} /> : <ChevronUp size={18} />}
            </button>

            <div
                className={`overflow-hidden transition-[max-height,opacity] duration-250 ease ${
                    collapsed ? 'max-h-0 opacity-0' : 'max-h-[100px] opacity-100'
                }`}
            >
                <div className="flex flex-wrap gap-4 mt-1 text-sm">
                    <div className="flex items-center gap-1">
                        <span className="text-muted">Step:</span>
                        {stepPeriods.map((p) => (
                            <button
                                key={p}
                                className={`px-1 rounded transition ${
                                    p === currentStep ? 'color-fg font-medium' : 'color-fg-secondary hover:color-fg'
                                }`}
                                onClick={() => setStep(p)}
                                disabled={p === currentStep}
                            >
                                {p}
                            </button>
                        ))}
                    </div>

                    <div className="flex items-center gap-1">
                        <span className="text-muted">Range:</span>
                        {rangePeriods.map((r) => (
                            <button
                                key={r}
                                className={`px-1 rounded transition ${
                                    r === currentRange ? 'color-fg font-medium' : 'color-fg-secondary hover:color-fg'
                                }`}
                                onClick={() => setRange(r)}
                                disabled={r === currentRange}
                            >
                                {r}
                            </button>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
}
