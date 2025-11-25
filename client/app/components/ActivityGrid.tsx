import { useQuery } from "@tanstack/react-query";
import {
  getActivity,
  type getActivityArgs,
  type ListenActivityItem,
} from "api/api";
import Popup from "./Popup";
import { useState } from "react";
import { useTheme } from "~/hooks/useTheme";
import ActivityOptsSelector from "./ActivityOptsSelector";
import type { Theme } from "~/styles/themes.css";

function getPrimaryColor(theme: Theme): string {
  const value = theme.primary;
  const rgbMatch = value.match(
    /^rgb\(\s*(\d{1,3})\s*,\s*(\d{1,3})\s*,\s*(\d{1,3})\s*\)$/
  );
  if (rgbMatch) {
    const [, r, g, b] = rgbMatch.map(Number);
    return "#" + [r, g, b].map((n) => n.toString(16).padStart(2, "0")).join("");
  }

  return value;
}
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

export default function ActivityGrid({
  step = "day",
  range = 182,
  month = 0,
  year = 0,
  artistId = 0,
  albumId = 0,
  trackId = 0,
  configurable = false,
}: Props) {
  const [stepState, setStep] = useState(step);
  const [rangeState, setRange] = useState(range);

  const { isPending, isError, data, error } = useQuery({
    queryKey: [
      "listen-activity",
      {
        step: stepState,
        range: rangeState,
        month: month,
        year: year,
        artist_id: artistId,
        album_id: albumId,
        track_id: trackId,
      },
    ],
    queryFn: ({ queryKey }) => getActivity(queryKey[1] as getActivityArgs),
  });

  const { theme, themeName } = useTheme();
  const color = getPrimaryColor(theme);

  if (isPending) {
    return (
      <div className="w-[500px]">
        <h2>Activity</h2>
        <p>Loading...</p>
      </div>
    );
  } else if (isError) {
    return (
      <div className="w-[500px]">
        <h2>Activity</h2>
        <p className="error">Error: {error.message}</p>
      </div>
    );
  }

  // from https://css-tricks.com/snippets/javascript/lighten-darken-color/
  function LightenDarkenColor(hex: string, lum: number) {
    // validate hex string
    hex = String(hex).replace(/[^0-9a-f]/gi, "");
    if (hex.length < 6) {
      hex = hex[0] + hex[0] + hex[1] + hex[1] + hex[2] + hex[2];
    }
    lum = lum || 0;

    // convert to decimal and change luminosity
    var rgb = "#",
      c,
      i;
    for (i = 0; i < 3; i++) {
      c = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
      c = Math.round(Math.min(Math.max(0, c + c * lum), 255)).toString(16);
      rgb += ("00" + c).substring(c.length);
    }

    return rgb;
  }

  const getDarkenAmount = (v: number, t: number): number => {
    // really ugly way to just check if this is for all items and not a specific item.
    // is it jsut better to just pass the target in as a var? probably.
    const adjustment =
      artistId == albumId && albumId == trackId && trackId == 0 ? 10 : 1;

    // automatically adjust the target value based on step
    // the smartest way to do this would be to have the api return the
    // highest value in the range. too bad im not smart
    switch (stepState) {
      case "day":
        t = 10 * adjustment;
        break;
      case "week":
        t = 20 * adjustment;
        break;
      case "month":
        t = 50 * adjustment;
        break;
      case "year":
        t = 100 * adjustment;
        break;
    }

    v = Math.min(v, t);
    if (themeName === "pearl") {
      // special case for the only light theme lol
      // could be generalized by pragmatically comparing the
      // lightness of the bg vs the primary but eh
      return (t - v) / t;
    } else {
      return ((v - t) / t) * 0.8;
    }
  };

  const CHUNK_SIZE = 26 * 7;
  const chunks = [];

  for (let i = 0; i < data.length; i += CHUNK_SIZE) {
    chunks.push(data.slice(i, i + CHUNK_SIZE));
  }

  return (
    <div className="flex flex-col items-start">
      <h2>Activity</h2>
      {configurable ? (
        <ActivityOptsSelector
          rangeSetter={setRange}
          currentRange={rangeState}
          stepSetter={setStep}
          currentStep={stepState}
        />
      ) : null}

      {chunks.map((chunk, index) => (
        <div
          key={index}
          className="w-auto grid grid-flow-col grid-rows-7 gap-[3px] md:gap-[5px] mb-4"
        >
          {chunk.map((item) => (
            <div
              key={new Date(item.start_time).toString()}
              className="w-[10px] sm:w-[12px] h-[10px] sm:h-[12px]"
            >
              <Popup
                position="top"
                space={12}
                extraClasses="left-2"
                inner={`${new Date(item.start_time).toLocaleDateString()} ${
                  item.listens
                } plays`}
              >
                <div
                  style={{
                    display: "inline-block",
                    background:
                      item.listens > 0
                        ? LightenDarkenColor(
                            color,
                            getDarkenAmount(item.listens, 100)
                          )
                        : "var(--color-bg-secondary)",
                  }}
                  className={`w-[10px] sm:w-[12px] h-[10px] sm:h-[12px] rounded-[2px] md:rounded-[3px] ${
                    item.listens > 0
                      ? ""
                      : "border-[0.5px] border-(--color-bg-tertiary)"
                  }`}
                ></div>
              </Popup>
            </div>
          ))}
        </div>
      ))}
    </div>
  );
}
