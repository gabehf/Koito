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
    const stepPeriods = ['day', 'week', 'month', 'year'];
    const rangePeriods = [105, 182, 365];

    const [collapsed, setCollapsed] = useState(true);

    const stepDisplay = (str: string): string =>
        str.split('_').map(w =>
            w.split('').map((char, index) =>
                index === 0 ? char.toUpperCase() : char).join('')
        ).join(' ');

    const rangeDisplay = (r: number): string => `${r}`;

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
            const cachedRange = parseInt(localStorage.getItem('activity_range_' + window.location.pathname.split('/')[1]) ?? '35');
            if (cachedRange) rangeSetter(cachedRange);
            const cachedStep = localStorage.getItem('activity_step_' + window.location.pathname.split('/')[1]);
            if (cachedStep) stepSetter(cachedStep);
        }
    }, []);

    return (
        <div className="flex flex-col gap-2 relative">
            <button
                onClick={() => setCollapsed(!collapsed)}
                className="text-sm underline self-start color-fg-secondary hover:color-fg transition absolute -top-9 left-20"
            >
                {collapsed ? <ChevronDown size={18} /> : <ChevronUp size={18} />}
            </button>

            {!collapsed && (
                <>
                    <div className="flex gap-2 items-center">
                        <p>Step:</p>
                        {stepPeriods.map((p, i) => (
                            <div key={`step_selector_${p}`}>
                                <button
                                    className={`period-selector ${p === currentStep ? 'color-fg' : 'color-fg-secondary'} ${i !== stepPeriods.length - 1 ? 'pr-2' : ''}`}
                                    onClick={() => setStep(p)}
                                    disabled={p === currentStep}
                                >
                                    {stepDisplay(p)}
                                </button>
                                <span className="color-fg-secondary">
                                    {i !== stepPeriods.length - 1 ? '|' : ''}
                                </span>
                            </div>
                        ))}
                    </div>

                    <div className="flex gap-2 items-center">
                        <p>Range:</p>
                        {rangePeriods.map((r, i) => (
                            <div key={`range_selector_${r}`}>
                                <button
                                    className={`period-selector ${r === currentRange ? 'color-fg' : 'color-fg-secondary'} ${i !== rangePeriods.length - 1 ? 'pr-2' : ''}`}
                                    onClick={() => setRange(r)}
                                    disabled={r === currentRange}
                                >
                                    {rangeDisplay(r)}
                                </button>
                                <span className="color-fg-secondary">
                                    {i !== rangePeriods.length - 1 ? '|' : ''}
                                </span>
                            </div>
                        ))}
                    </div>
                </>
            )}
        </div>
    );
}
