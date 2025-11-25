import Timeframe from "~/types/timeframe";

const timeframeToInterval = (timeframe: Timeframe): string => {
  switch (timeframe) {
    case Timeframe.Day:
      return "1 day";
    case Timeframe.Week:
      return "1 week";
    case Timeframe.Month:
      return "1 month";
    case Timeframe.Year:
      return "1 year";
    case Timeframe.AllTime:
      return "99 years";
  }
};

function timeSince(date: Date) {
  const now = new Date();
  const seconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  const intervals = [
    { label: "year", seconds: 31536000 },
    { label: "month", seconds: 2592000 },
    { label: "week", seconds: 604800 },
    { label: "day", seconds: 86400 },
    { label: "hour", seconds: 3600 },
    { label: "minute", seconds: 60 },
    { label: "second", seconds: 1 },
  ];

  for (const interval of intervals) {
    const count = Math.floor(seconds / interval.seconds);
    if (count >= 1) {
      return `${count} ${interval.label}${count !== 1 ? "s" : ""} ago`;
    }
  }

  return "just now";
}

export { timeSince };

type hsl = {
  h: number;
  s: number;
  l: number;
};

const hexToHSL = (hex: string): hsl => {
  let r = 0,
    g = 0,
    b = 0;
  hex = hex.replace("#", "");

  if (hex.length === 3) {
    r = parseInt(hex[0] + hex[0], 16);
    g = parseInt(hex[1] + hex[1], 16);
    b = parseInt(hex[2] + hex[2], 16);
  } else if (hex.length === 6) {
    r = parseInt(hex.substring(0, 2), 16);
    g = parseInt(hex.substring(2, 4), 16);
    b = parseInt(hex.substring(4, 6), 16);
  }

  r /= 255;
  g /= 255;
  b /= 255;

  const max = Math.max(r, g, b),
    min = Math.min(r, g, b);
  let h = 0,
    s = 0,
    l = (max + min) / 2;

  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    switch (max) {
      case r:
        h = (g - b) / d + (g < b ? 6 : 0);
        break;
      case g:
        h = (b - r) / d + 2;
        break;
      case b:
        h = (r - g) / d + 4;
        break;
    }
    h /= 6;
  }

  return {
    h: Math.round(h * 360),
    s: Math.round(s * 100),
    l: Math.round(l * 100),
  };
};

const timeListenedString = (seconds: number) => {
  if (!seconds) return "";

  let minutes = Math.floor(seconds / 60);
  return `${minutes} minutes listened`;
};

export { hexToHSL, timeListenedString };
export type { hsl };
