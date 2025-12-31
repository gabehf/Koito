import { imageUrl, type Artist } from "api/api";

interface args {
  title?: string;
  name?: string;
  image: string;
  minutes_listened: number;
  time_listened: number;
  artists?: Artist;
}

export default function RewindTopItem(args: args[]) {
  console.log(args);
  if (args === undefined || args.length < 1) {
    return <></>;
  }
  const img = imageUrl(args[0].image, "medium");

  return (
    <div className="flex gap-2">
      <div className="rewind-top-item-image">
        <img src={img} />
      </div>
      <div className="flex flex-col gap-1">
        <h3>{args[0].title || args[0].name}</h3>
        {args.map((e) => (
          <div className="">{e.title || e.name}</div>
        ))}
      </div>
    </div>
  );
}
