import type { Ranked } from "api/api";
import Image from "../primitives/Image";

type TopItemProps<T> = {
  title: string;
  imageSrc: string;
  items: Ranked<T>[];
  getLabel: (item: T) => string;
  includeTime?: boolean;
};

export function RewindTopItem<
  T extends {
    id: string | number;
    listen_count: number;
    time_listened: number;
  },
>({ title, imageSrc, items, getLabel, includeTime }: TopItemProps<T>) {
  const [top, ...rest] = items;

  if (!top) return null;

  const titlesizeclass =
    getLabel(top.item).length > 28
      ? "text-xl lg:text-2xl"
      : "text-2xl lg:text-4xl";

  // const titlesizeclass = true ? "text-xl lg:text-2xl" : "text-2xl lg:text-4xl";

  return (
    <div className="flex flex-col sm:flex-row gap-5">
      <div className="rewind-top-item-image min-w-[210px]">
        <Image size={210} src={imageSrc} alt={title} />
      </div>

      <div className="flex flex-col gap-1">
        <h5 className="-mb-1 uppercase text-(--color-fg-secondary)">{title}</h5>

        <div className="flex items-center gap-2">
          <div className="flex flex-col items-start mb-2">
            <h5
              className={`${titlesizeclass} font-semibold mb-2 mt-2 max-w-100 wrap-normal`}
            >
              {getLabel(top.item)}
            </h5>
            <span className="text-(--color-fg-secondary) -mt-2 text-sm">
              {`${top.item.listen_count} plays`}
              {includeTime
                ? ` (${Math.floor(top.item.time_listened / 60)} minutes)`
                : ``}
            </span>
          </div>
        </div>

        {rest.map((e) => (
          <div key={e.item.id} className="text-sm max-w-90 wrap-normal">
            {getLabel(e.item)}
            <span className="text-(--color-fg-secondary)">
              {` - ${e.item.listen_count} plays`}
              {includeTime
                ? ` (${Math.floor(e.item.time_listened / 60)} minutes)`
                : ``}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
