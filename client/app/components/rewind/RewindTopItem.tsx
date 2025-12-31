type TopItemProps<T> = {
  title: string;
  imageSrc: string;
  items: T[];
  getLabel: (item: T) => string;
  includeTime?: boolean;
};

export function RewindTopItem<
  T extends {
    id: string | number;
    listen_count: number;
    time_listened: number;
  }
>({ title, imageSrc, items, getLabel, includeTime }: TopItemProps<T>) {
  const [top, ...rest] = items;

  if (!top) return null;

  return (
    <div className="flex gap-5">
      <div className="rewind-top-item-image">
        <img className="w-50 h-50" src={imageSrc} />
      </div>

      <div className="flex flex-col gap-1">
        <h4 className="-mb-1">{title}</h4>

        <div className="flex items-center gap-2">
          <div className="flex flex-col items-start mb-2">
            <h2>{getLabel(top)}</h2>
            <span className="text-(--color-fg-tertiary) -mt-3 text-sm">
              {`${top.listen_count} plays`}
              {includeTime
                ? ` (${Math.floor(top.time_listened / 60)} minutes)`
                : ``}
            </span>
          </div>
        </div>

        {rest.map((e) => (
          <div key={e.id} className="text-sm">
            {getLabel(e)}
            <span className="text-(--color-fg-tertiary)">
              {` - ${e.listen_count} plays`}
              {includeTime
                ? ` (${Math.floor(e.time_listened / 60)} minutes)`
                : ``}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
