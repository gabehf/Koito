import { useQuery } from "@tanstack/react-query";
import Popup from "./Popup";
import { apiFetch, type ListenActivityItem } from "api/api";
import CardHeader from "./primitives/CardHeader";
import useWindowWidth from "~/hooks/useWindowWidth";

interface Props {
  step?: string;
  range?: number;
  month?: number;
  year?: number;
  artistId?: number;
  albumId?: number;
  trackId?: number;
  configurable?: boolean;
  autoAdjust?: boolean;
}

const getActivity = (args: {
  step: string;
  range: number;
  month: number;
  year: number;
  artist_id: number;
  album_id: number;
  track_id: number;
}) => apiFetch<ListenActivityItem[]>("/apis/web/v1/listen-activity", args);

const NUM_WEEKS = 36;

export default function ActivityGrid({
  month = 0,
  year = 0,
  artistId = 0,
  albumId = 0,
  trackId = 0,
}: Props) {
  const args = {
    step: "day",
    range: NUM_WEEKS * 7,
    month,
    year,
    artist_id: artistId,
    album_id: albumId,
    track_id: trackId,
  };
  const { isPending, isError, data, error } = useQuery({
    queryKey: ["listen-activity", args],
    queryFn: () => getActivity(args),
  });

  const width = useWindowWidth();

  const header = "Activity";

  if (isPending) {
    return (
      <div className="w-[350px]">
        <CardHeader>{header}</CardHeader>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[350px]">
        <CardHeader>{header}</CardHeader>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  let target = 100;
  if (trackId !== 0) {
    target = 5;
  } else if (albumId !== 0) {
    target = 10;
  } else if (artistId !== 0) {
    target = 20;
  }

  const getBlendAmount = (v: number, t: number): number => {
    v = Math.min(v, t);
    return 0.1 + (v / t) * 0.9;
  };

  // Build a lookup from normalized date key → listen count
  const listenMap = new Map<string, number>();
  for (const item of data) {
    const d = new Date(item.start_time);
    const key = `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}`;
    listenMap.set(key, item.listens);
  }

  // Align the grid to calendar weeks (Monday = row 0, Sunday = row 6).
  // Column 0 is the oldest week; the last column is the current (partial) week.
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const daysSinceMonday = (today.getDay() + 6) % 7; // Mon=0 … Sun=6
  const gridStart = new Date(today);
  gridStart.setDate(
    gridStart.getDate() - daysSinceMonday - (NUM_WEEKS - 1) * 7
  );

  const cells: { date: Date; listens: number; isFuture: boolean }[] = [];
  for (let col = 0; col < NUM_WEEKS; col++) {
    for (let row = 0; row < 7; row++) {
      const date = new Date(gridStart);
      date.setDate(date.getDate() + col * 7 + row);
      const isFuture = date > today;
      const key = `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
      cells.push({ date, listens: listenMap.get(key) ?? 0, isFuture });
    }
  }

  // if window size is small (640px)
  if (width <= 640) {
    // remove 10 weeks from data
    cells.splice(0, 10 * 7 - 1);
  }

  const CELL_W = "w-[9px] sm:w-[10px]";
  const CELL_H = "h-[9px] sm:h-[10px]";
  const CELL_GAP = "gap-[2px] md:gap-[3px]";
  const CELL_RADIUS = "rounded-[2px]";
  const DAY_LABELS = ["Mon", "", "Wed", "", "Fri", "", "Sun"];

  return (
    <div className="flex flex-col items-start">
      <CardHeader isOffset>{header}</CardHeader>
      <div className="flex flex-col items-center gap-6 pt-4 sm:pt-6 px-4 sm:px-7 pb-4 sm:pb-5 border bg-(--color-bg-secondary) rounded-(--border-radius)">
        <div className="flex items-start gap-2">
          <div className={`grid grid-rows-7 ${CELL_GAP}`}>
            {DAY_LABELS.map((label, i) => (
              <div
                key={i}
                className={`${CELL_H} flex items-center justify-end`}
              >
                <span className="text-[8px] sm:text-[11px] mt-4 leading-none text-(--color-fg-secondary)">
                  {label}
                </span>
              </div>
            ))}
          </div>
          <div className={`w-auto grid grid-flow-col grid-rows-7 ${CELL_GAP}`}>
            {cells.map((cell) => (
              <div
                key={cell.date.toISOString()}
                className={`${CELL_W} ${CELL_H}`}
              >
                {cell.isFuture ? (
                  <div className={`${CELL_W} ${CELL_H} ${CELL_RADIUS}`} />
                ) : (
                  <Popup
                    position="top"
                    space={12}
                    extraClasses="left-2"
                    inner={`${cell.date.toLocaleDateString()} ${
                      cell.listens
                    } plays`}
                  >
                    <div
                      style={{
                        display: "inline-block",
                        background:
                          cell.listens > 0
                            ? "var(--color-primary)"
                            : "var(--color-bg-secondary)",
                        opacity: getBlendAmount(cell.listens, target),
                      }}
                      className={`${CELL_W} ${CELL_H} ${CELL_RADIUS} ${
                        cell.listens > 0
                          ? ""
                          : "border-[0.5px] border-(--color-bg-tertiary)"
                      }`}
                    />
                  </Popup>
                )}
              </div>
            ))}
          </div>
        </div>
        <div className="flex justify-around w-full">
          <div className="text-[11px] color-fg-secondary">
            Streak · <span className="text-(--color-primary)">234</span>
          </div>
          <div className="flex items-center gap-2 text-[11px] color-fg-secondary">
            <div>Less</div>
            <div className="grid grid-cols-5 gap-1">
              {[1, 2, 3, 4, 5].map((i) => (
                <div
                  style={{
                    display: "inline-block",
                    background: "var(--color-primary)",
                    opacity: getBlendAmount(i, 5),
                  }}
                  className={`w-[8px] sm:w-[9px] h-[8px] sm:h-[9px] rounded-[3px] ${"border-[0.5px] border-(--color-bg-tertiary)"}`}
                />
              ))}
            </div>
            <div>More</div>
          </div>
        </div>
      </div>
    </div>
  );
}
