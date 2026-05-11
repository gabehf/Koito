import CardHeader from "../primitives/CardHeader";
import { MediaItemSkeleton } from "../primitives/MediaItem";

interface TopCardSkeletonProps {
  header: string;
  numItems?: number;
}

export function TopCardSkeleton({
  header,
  numItems = 5,
}: TopCardSkeletonProps) {
  return (
    <div>
      <CardHeader isOffset>{header}</CardHeader>
      <div className="max-w-[350px] card animate-pulse">
        <div
          className="relative bg"
          style={{
            width: 348,
            height: 350,
            borderRadius: "var(--border-radius) var(--border-radius) 0 0",
            maskImage: "linear-gradient(to bottom, black, transparent)",
            WebkitMaskImage: "linear-gradient(to bottom, black, transparent)",
          }}
        >
          <div className="absolute bottom-10 left-5 flex flex-col gap-2">
            <div className="h-7 w-48 bg animate-pulse rounded-(--border-radius)" />
            <div className="h-4 w-32 bg animate-pulse rounded-(--border-radius)" />
            <div className="h-3 w-16 bg animate-pulse rounded-(--border-radius)" />
          </div>
        </div>
        <div className="flex flex-col items-start">
          {Array.from({ length: numItems - 1 }).map((_, i) => (
            <div className="px-6 pb-6" key={i}>
              <MediaItemSkeleton bgColor="bg" size="md" subtitle meta />
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
