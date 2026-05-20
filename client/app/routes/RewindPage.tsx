import Rewind from "~/components/rewind/Rewind";
import { type RewindStats } from "api/api";
import { useState } from "react";
import type { LoaderFunctionArgs } from "react-router";
import { useLoaderData, useLocation, useNavigate } from "react-router";
import { getRewindParams } from "~/utils/utils";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { useAppContext } from "~/providers/AppProvider";

const months = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December",
];

const parseParams = (
  params: URLSearchParams,
): { year: number; month?: number } => {
  const rawYear = params.get("year");
  const rawMonth = params.get("month");

  if (!rawYear && !rawMonth) return getRewindParams();
  if (rawYear && rawMonth)
    return { year: parseInt(rawYear), month: parseInt(rawMonth) };
  if (rawYear) return { year: parseInt(rawYear) };

  const month = parseInt(rawMonth!);
  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonth = now.getMonth() + 1;
  return { year: month <= currentMonth ? currentYear : currentYear - 1, month };
};

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const { year, month } = parseParams(url.searchParams);

  const res = await fetch(`/apis/web/v1/summary?year=${year}&month=${month}`);
  if (!res.ok) throw new Response("Failed to load summary", { status: 500 });

  const stats: RewindStats = await res.json();
  stats.title = `Your ${month && month > 0 ? months[month - 1] : ""} ${year} Rewind`;
  return { stats };
}

export default function RewindPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const { stats } = useLoaderData<{ stats: RewindStats }>();
  const { firstActivity } = useAppContext();

  const { year: loadedYear, month: loadedMonth } = parseParams(
    new URLSearchParams(location.search),
  );

  const [year, setYear] = useState(loadedYear);
  const [month, setMonth] = useState(loadedMonth);
  const [showYearly, setShowYearly] = useState(
    !loadedMonth || loadedMonth === 0,
  );
  const [showTime, setShowTime] = useState(false);

  const updateParams = (updates: Record<string, string | null>) => {
    const nextParams = new URLSearchParams(location.search);
    for (const [key, val] of Object.entries(updates)) {
      val !== null ? nextParams.set(key, val) : nextParams.delete(key);
    }
    navigate(`/rewind?${nextParams.toString()}`, { replace: false });
  };

  const navigateMonth = (newMonth: number) => {
    setMonth(newMonth);
    updateParams({ year: year.toString(), month: newMonth.toString() });
  };

  const navigateYear = (direction: "prev" | "next") => {
    setYear(direction === "next" ? year + 1 : year - 1);
  };

  // month is 0-indexed (array index), matching JS Date convention
  function isFuture(month: number, year: number) {
    return new Date(year, month + 1) > new Date();
  }

  function isBeforeFirstActivity(month: number, year: number) {
    return !!firstActivity && new Date(year, month + 1) < firstActivity;
  }

  function isUnavailable(monthIndex: number, year: number) {
    return (
      isFuture(monthIndex, year) || isBeforeFirstActivity(monthIndex, year)
    );
  }

  const pgTitle = `${stats.title} - Koito`;

  return (
    <div className="w-full min-h-screen">
      <title>{pgTitle}</title>
      <meta property="og:title" content={pgTitle} />
      <meta name="description" content={pgTitle} />
      <div className="flex flex-col lg:flex-row items-center lg:items-start lg:mt-15 mt-5 gap-10 w-19/20 mx-auto px-5 md:px-10">
        <div className="flex flex-col items-start gap-4">
          <div className="flex flex-col items-center gap-4 mt-4 sm:mt-14 w-[250px] mx-auto">
            {!showYearly && (
              <>
                <div className="flex items-center gap-6 justify-around">
                  <button
                    onClick={() => navigateYear("prev")}
                    className="p-2 disabled:opacity-0"
                    disabled={
                      new Date(year - 1, month || 0) > new Date() ||
                      (firstActivity && year - 1 < firstActivity.getFullYear())
                    }
                  >
                    <ChevronLeft size={20} />
                  </button>
                  <p className="font-medium text-xl text-center w-30">{year}</p>
                  <button
                    onClick={() => navigateYear("next")}
                    className="p-2 disabled:opacity-0"
                    disabled={isFuture(0, year + 1)}
                  >
                    <ChevronRight size={20} />
                  </button>
                </div>
                <div className="grid grid-cols-2 gap-2">
                  {months.map((m, i) => {
                    const isSelected =
                      month !== undefined &&
                      i === month - 1 &&
                      loadedYear === year;
                    const unavailable = isUnavailable(i, year);
                    return (
                      <button
                        key={m}
                        className={`px-12 py-2 rounded-(--border-radius)
                          ${isSelected ? "card" : "border-1 border-(--color-bg)"}
                          ${unavailable ? "color-fg-tertiary" : "hover:bg-(--color-bg-secondary)"}
                        `}
                        onClick={() => navigateMonth(i + 1)}
                        disabled={unavailable || isSelected}
                      >
                        {m.substring(0, 3)}
                      </button>
                    );
                  })}
                </div>
              </>
            )}
            {showYearly && firstActivity && (
              <div className="grid grid-cols-2 gap-2">
                {Array.from(
                  {
                    length:
                      new Date().getFullYear() - firstActivity.getFullYear(),
                  },
                  (_, index) => firstActivity.getFullYear() + index,
                ).map((y) => (
                  <button
                    key={y}
                    className={`px-10 py-2 rounded-(--border-radius) hover:bg-(--color-bg-secondary)
                      ${y === year && !month ? "card" : "border-1 border-(--color-bg)"}
                    `}
                    onClick={() => {
                      setYear(y);
                      setMonth(undefined);
                      updateParams({ year: y.toString(), month: null });
                    }}
                    disabled={(!month || month === 0) && y === year}
                  >
                    {y}
                  </button>
                ))}
              </div>
            )}
            <div className="flex items-center gap-2">
              <input
                id="show-time-checkbox"
                type="checkbox"
                checked={showTime}
                onChange={(e) => setShowTime(e.target.checked)}
              />
              <label htmlFor="show-time-checkbox">Show time listened?</label>
            </div>
          </div>
        </div>
        <div className="flex flex-col items-center gap-4">
          <div className="flex justify-around w-2/3 mb-4">
            <button
              className="period-selector"
              onClick={() => setShowYearly(false)}
              disabled={!showYearly}
            >
              MONTHLY
            </button>
            <button
              className="period-selector"
              onClick={() => setShowYearly(true)}
              disabled={showYearly}
            >
              YEARLY
            </button>
          </div>
          {stats && <Rewind stats={stats} includeTime={showTime} />}
        </div>
      </div>
    </div>
  );
}
