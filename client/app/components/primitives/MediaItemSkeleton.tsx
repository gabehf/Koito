interface SkeletonProps {
  size: "sm" | "md" | "lg";
  subtitle?: boolean;
  meta?: boolean;
  className?: string;
  bgColor?: string;
  alignTop?: boolean;
}

export function MediaItemSkeleton({
  size,
  subtitle,
  meta,
  className,
  bgColor,
  alignTop,
}: SkeletonProps) {
  const sizeToPx = (size: "sm" | "md" | "lg") => {
    switch (size) {
      case "sm":
        return 56;
      case "md":
        return 90;
      case "lg":
        return 125;
    }
  };
  const px = sizeToPx(size);
  const titleW = size === "sm" ? 100 : size === "md" ? 140 : 180;
  const bgClass = bgColor || "bg-secondary";

  return (
    <div
      className={`flex ${alignTop ? "items-start" : "items-center"} gap-3 ${
        className ?? ""
      }`}
    >
      <div
        className={`rounded-(--border-radius) ${bgClass} animate-pulse shrink-0`}
        style={{ width: px, height: px }}
      />
      <div
        className="flex flex-col items-start gap-2"
        style={alignTop ? { marginTop: 6 } : undefined}
      >
        <div
          className={`h-4 ${bgClass} animate-pulse rounded-(--border-radius)`}
          style={{ width: titleW }}
        />
        {subtitle && (
          <div
            className={`h-3 ${bgClass} animate-pulse rounded-(--border-radius)`}
            style={{ width: titleW * 0.7 }}
          />
        )}
        {meta && (
          <div
            className={`h-3 ${bgClass} animate-pulse rounded-(--border-radius)`}
            style={{ width: titleW * 0.5 }}
          />
        )}
      </div>
    </div>
  );
}
