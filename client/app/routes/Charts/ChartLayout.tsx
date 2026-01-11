import { useFetcher, useLocation, useNavigate } from "react-router";
import { useEffect, useState } from "react";
import { average } from "color.js";
import { imageUrl, type PaginatedResponse } from "api/api";
import PeriodSelector from "~/components/PeriodSelector";

interface ChartLayoutProps<T> {
  title: "Top Albums" | "Top Tracks" | "Top Artists" | "Last Played";
  initialData: PaginatedResponse<T>;
  endpoint: string;
  render: (opts: {
    data: PaginatedResponse<T>;
    page: number;
    onNext: () => void;
    onPrev: () => void;
  }) => React.ReactNode;
}

export default function ChartLayout<T>({
  title,
  initialData,
  endpoint,
  render,
}: ChartLayoutProps<T>) {
  const pgTitle = `${title} - Koito`;

  const fetcher = useFetcher();
  const location = useLocation();
  const navigate = useNavigate();

  const currentParams = new URLSearchParams(location.search);
  const currentPage = parseInt(currentParams.get("page") || "1", 10);

  const data: PaginatedResponse<T> = fetcher.data?.[endpoint]
    ? fetcher.data[endpoint]
    : initialData;

  const [bgColor, setBgColor] = useState<string>("(--color-bg)");

  useEffect(() => {
    if ((data?.items?.length ?? 0) === 0) return;

    const img = (data.items[0] as any)?.image;
    if (!img) return;

    average(imageUrl(img, "small"), { amount: 1 }).then((color) => {
      setBgColor(`rgba(${color[0]},${color[1]},${color[2]},0.4)`);
    });
  }, [data]);

  const period = currentParams.get("period") ?? "day";
  const year = currentParams.get("year");
  const month = currentParams.get("month");
  const week = currentParams.get("week");

  const updateParams = (params: Record<string, string | null>) => {
    const nextParams = new URLSearchParams(location.search);

    for (const key in params) {
      const val = params[key];
      if (val !== null) {
        nextParams.set(key, val);
      } else {
        nextParams.delete(key);
      }
    }

    const url = `/${endpoint}?${nextParams.toString()}`;
    navigate(url, { replace: false });
  };

  const handleSetPeriod = (p: string) => {
    updateParams({
      period: p,
      page: "1",
      year: null,
      month: null,
      week: null,
    });
  };
  const handleSetYear = (val: string) => {
    if (val == "") {
      updateParams({
        period: period,
        page: "1",
        year: null,
        month: null,
        week: null,
      });
      return;
    }
    updateParams({
      period: null,
      page: "1",
      year: val,
    });
  };
  const handleSetMonth = (val: string) => {
    updateParams({
      period: null,
      page: "1",
      year: year ?? new Date().getFullYear().toString(),
      month: val,
    });
  };
  const handleSetWeek = (val: string) => {
    updateParams({
      period: null,
      page: "1",
      year: year ?? new Date().getFullYear().toString(),
      month: null,
      week: val,
    });
  };

  useEffect(() => {
    fetcher.load(`/${endpoint}?${currentParams.toString()}`);
  }, [location.search]);

  const setPage = (nextPage: number) => {
    const nextParams = new URLSearchParams(location.search);
    nextParams.set("page", String(nextPage));
    const url = `/${endpoint}?${nextParams.toString()}`;
    fetcher.load(url);
    navigate(url, { replace: false });
  };

  const handleNextPage = () => setPage(currentPage + 1);
  const handlePrevPage = () => setPage(currentPage - 1);

  const yearOptions = Array.from(
    { length: 10 },
    (_, i) => `${new Date().getFullYear() - i}`
  );
  const monthOptions = Array.from({ length: 12 }, (_, i) => `${i + 1}`);
  const weekOptions = Array.from({ length: 53 }, (_, i) => `${i + 1}`);

  const getDateRange = (): string => {
    let from: Date;
    let to: Date;

    const now = new Date();
    const currentYear = now.getFullYear();
    const currentMonth = now.getMonth(); // 0-indexed
    const currentDate = now.getDate();

    if (year && month) {
      from = new Date(parseInt(year), parseInt(month) - 1, 1);
      to = new Date(from);
      to.setMonth(from.getMonth() + 1);
      to.setDate(0);
    } else if (year && week) {
      const base = new Date(parseInt(year), 0, 1); // Jan 1 of the year
      const weekNumber = parseInt(week);
      from = new Date(base);
      from.setDate(base.getDate() + (weekNumber - 1) * 7);
      to = new Date(from);
      to.setDate(from.getDate() + 6);
    } else if (year) {
      from = new Date(parseInt(year), 0, 1);
      to = new Date(parseInt(year), 11, 31);
    } else {
      switch (period) {
        case "day":
          from = new Date(now);
          to = new Date(now);
          break;
        case "week":
          to = new Date(now);
          from = new Date(now);
          from.setDate(to.getDate() - 6);
          break;
        case "month":
          to = new Date(now);
          from = new Date(now);
          if (currentMonth === 0) {
            from = new Date(currentYear - 1, 11, currentDate);
          } else {
            from = new Date(currentYear, currentMonth - 1, currentDate);
          }
          break;
        case "year":
          to = new Date(now);
          from = new Date(currentYear - 1, currentMonth, currentDate);
          break;
        case "all_time":
          return "All Time";
        default:
          return "";
      }
    }

    const formatter = new Intl.DateTimeFormat(undefined, {
      year: "numeric",
      month: "long",
      day: "numeric",
    });

    return `${formatter.format(from)} - ${formatter.format(to)}`;
  };

  return (
    <div
      className="w-full min-h-screen"
      style={{
        background: `linear-gradient(to bottom, ${bgColor}, var(--color-bg) 500px)`,
        transition: "1000",
      }}
    >
      <title>{pgTitle}</title>
      <meta property="og:title" content={pgTitle} />
      <meta name="description" content={pgTitle} />
      <div className="w-19/20 sm:17/20 mx-auto pt-6 sm:pt-12">
        <h1>{title}</h1>
        <div className="flex flex-col items-start md:flex-row sm:items-center gap-4">
          <PeriodSelector
            current={period}
            setter={handleSetPeriod}
            disableCache
          />
          <div className="flex gap-5">
            <select
              value={year ?? ""}
              onChange={(e) => handleSetYear(e.target.value)}
              className="px-2 py-1 rounded border border-gray-400"
            >
              <option value="">Year</option>
              {yearOptions.map((y) => (
                <option key={y} value={y}>
                  {y}
                </option>
              ))}
            </select>
            <select
              value={month ?? ""}
              onChange={(e) => handleSetMonth(e.target.value)}
              className="px-2 py-1 rounded border border-gray-400"
            >
              <option value="">Month</option>
              {monthOptions.map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
            <select
              value={week ?? ""}
              onChange={(e) => handleSetWeek(e.target.value)}
              className="px-2 py-1 rounded border border-gray-400"
            >
              <option value="">Week</option>
              {weekOptions.map((w) => (
                <option key={w} value={w}>
                  {w}
                </option>
              ))}
            </select>
          </div>
        </div>
        <p className="mt-2 text-sm text-color-fg-secondary">{getDateRange()}</p>
        <div className="mt-10 sm:mt-20 flex mx-auto justify-between">
          {render({
            data,
            page: currentPage,
            onNext: handleNextPage,
            onPrev: handlePrevPage,
          })}
        </div>
      </div>
    </div>
  );
}
