import Rewind from "~/components/rewind/Rewind";
import type { Route } from "./+types/Home";
import { imageUrl, type RewindStats } from "api/api";
import { useEffect, useState } from "react";
import type { LoaderFunctionArgs } from "react-router";
import { useLoaderData } from "react-router";
import { getRewindParams } from "~/utils/utils";
import { useNavigate } from "react-router";
import { average } from "color.js";
import { ChevronLeft, ChevronRight } from "lucide-react";

// TODO: Bind year and month selectors to what data actually exists

const months = [
  "Full Year",
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

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const year = parseInt(
    url.searchParams.get("year") || getRewindParams().year.toString()
  );
  const month = parseInt(
    url.searchParams.get("month") || getRewindParams().month.toString()
  );

  const res = await fetch(`/apis/web/v1/summary?year=${year}&month=${month}`);
  if (!res.ok) {
    throw new Response("Failed to load summary", { status: 500 });
  }

  const stats: RewindStats = await res.json();
  stats.title = `Your ${month === 0 ? "" : months[month]} ${year} Rewind`;
  return { stats };
}

export default function RewindPage() {
  const currentParams = new URLSearchParams(location.search);
  let year = parseInt(
    currentParams.get("year") || getRewindParams().year.toString()
  );
  let month = parseInt(
    currentParams.get("month") || getRewindParams().month.toString()
  );
  const navigate = useNavigate();
  const [showTime, setShowTime] = useState(false);
  const { stats: stats } = useLoaderData<{ stats: RewindStats }>();
  const latestRewindParams = getRewindParams();
  const rewindView = month === 0 ? "year" : "month";

  const [bgColor, setBgColor] = useState<string>("(--color-bg)");

  useEffect(() => {
    if (!stats.top_artists[0]) return;

    const img = (stats.top_artists[0] as any)?.item.image;
    if (!img) return;

    average(imageUrl(img, "small"), { amount: 1 }).then((color) => {
      setBgColor(`rgba(${color[0]},${color[1]},${color[2]},0.4)`);
    });
  }, [stats]);

  const updateParams = (params: Record<string, string | null>) => {
    const nextParams = new URLSearchParams(location.search);

    for (const key in params) {
      const val = params[key];

      if (val !== null) {
        nextParams.set(key, val);
      }
    }

    const url = `/rewind?${nextParams.toString()}`;

    navigate(url, { replace: false });
  };

  const getAdjacentMonthParams = (
    currentYear: number,
    currentMonth: number,
    direction: "prev" | "next"
  ) => {
    if (direction === "next") {
      if (currentMonth === 12) {
        return { year: currentYear + 1, month: 1 };
      }

      if (currentMonth === 0) {
        return { year: currentYear, month: 1 };
      }

      return { year: currentYear, month: currentMonth + 1 };
    }

    if (currentMonth === 1) {
      return { year: currentYear - 1, month: 12 };
    }

    if (currentMonth === 0) {
      return { year: currentYear - 1, month: 12 };
    }

    return { year: currentYear, month: currentMonth - 1 };
  };

  const nextMonthParams = getAdjacentMonthParams(year, month, "next");
  const disableNextMonth =
    nextMonthParams.year > latestRewindParams.year ||
    (nextMonthParams.year === latestRewindParams.year &&
      nextMonthParams.month > latestRewindParams.month);

  const navigateMonth = (direction: "prev" | "next") => {
    const nextParams = getAdjacentMonthParams(year, month, direction);

    updateParams({
      year: nextParams.year.toString(),
      month: nextParams.month.toString(),
    });
  };

  const setRewindView = (view: "year" | "month") => {
    if (view === "year") {
      updateParams({
        year: year.toString(),
        month: "0",
      });
      return;
    }

    const nextMonth = year === latestRewindParams.year ? latestRewindParams.month : 12;

    updateParams({
      year: year.toString(),
      month: nextMonth.toString(),
    });
  };

  const navigateYear = (direction: "prev" | "next") => {
    if (direction === "next") {
      year += 1;
    } else {
      year -= 1;
    }

    updateParams({
      year: year.toString(),
      month: month.toString(),
    });
  };

  const pgTitle = `${stats.title} - Koito`;

  return (
    <div
      className="w-full min-h-screen"
      style={{
        background: `linear-gradient(to bottom, ${bgColor}, var(--color-bg) 500px)`,
        transition: "1000",
      }}
    >
      <div className="flex flex-col items-start sm:items-center gap-4">
        <title>{pgTitle}</title>
        <meta property="og:title" content={pgTitle} />
        <meta name="description" content={pgTitle} />
        <div className="flex flex-col lg:flex-row items-start lg:mt-15 mt-5 gap-10 w-19/20 px-5 md:px-20">
          <div className="flex flex-col items-start gap-4">
            <div className="flex flex-col items-start gap-4 py-8">
              <div className="flex w-[15rem] items-center rounded-lg bg p-1">
                <button
                  onClick={() => setRewindView("year")}
                  className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors hover-bg-secondary ${
                    rewindView === "year"
                      ? "bg-secondary"
                      : "color-fg-secondary"
                  }`}
                >
                  Full Year
                </button>
                <button
                  onClick={() => setRewindView("month")}
                  className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors hover-bg-secondary ${
                    rewindView === "month"
                      ? "bg-secondary"
                      : "color-fg-secondary"
                  }`}
                >
                  Monthly
                </button>
              </div>
              {rewindView === "month" && (
                <div className="flex w-[15rem] items-center justify-between">
                  <button
                    onClick={() => navigateMonth("prev")}
                    className="p-2 disabled:text-(--color-fg-tertiary)"
                  >
                    <ChevronLeft size={20} />
                  </button>
                  <p className="font-medium text-xl text-center w-30">
                    {months[month]}
                  </p>
                  <button
                    onClick={() => navigateMonth("next")}
                    className="p-2 disabled:text-(--color-fg-tertiary)"
                    disabled={disableNextMonth}
                  >
                    <ChevronRight size={20} />
                  </button>
                </div>
              )}
              <div className="flex w-[15rem] items-center justify-between">
                <button
                  onClick={() => navigateYear("prev")}
                  className="p-2 disabled:text-(--color-fg-tertiary)"
                  disabled={new Date(year - 1, month) > new Date()}
                >
                  <ChevronLeft size={20} />
                </button>
                <p className="font-medium text-xl text-center w-30">{year}</p>
                <button
                  onClick={() => navigateYear("next")}
                  className="p-2 disabled:text-(--color-fg-tertiary)"
                  disabled={
                    // Next year date is in the future OR
                    new Date(year + 1, month - 1) > new Date() ||
                    // Next year date is current full year OR
                    (month == 0 && new Date().getFullYear() === year + 1) ||
                    // Next year date is current month
                    (new Date().getMonth() === month - 1 &&
                      new Date().getFullYear() === year + 1)
                  }
                >
                  <ChevronRight size={20} />
                </button>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <label htmlFor="show-time-checkbox">Show time listened?</label>
              <input
                type="checkbox"
                name="show-time-checkbox"
                checked={showTime}
                onChange={(e) => setShowTime(!showTime)}
              ></input>
            </div>
          </div>
          {stats !== undefined && (
            <Rewind stats={stats} includeTime={showTime} />
          )}
        </div>
      </div>
    </div>
  );
}
