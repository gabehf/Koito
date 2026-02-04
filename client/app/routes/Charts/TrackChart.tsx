import TopItemList from "~/components/TopItemList";
import ChartLayout from "./ChartLayout";
import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import { type Track, type PaginatedResponse, type Ranked } from "api/api";

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = url.searchParams.get("page") || "0";
  url.searchParams.set("page", page);

  const res = await fetch(
    `/apis/web/v1/top-tracks?${url.searchParams.toString()}`
  );
  if (!res.ok) {
    throw new Response("Failed to load top tracks", { status: 500 });
  }

  const top_tracks: PaginatedResponse<Track> = await res.json();
  return { top_tracks };
}

export default function TrackChart() {
  const { top_tracks: initialData } = useLoaderData<{
    top_tracks: PaginatedResponse<Ranked<Track>>;
  }>();

  return (
    <ChartLayout
      title="Top Tracks"
      initialData={initialData}
      endpoint="chart/top-tracks"
      render={({ data, page, onNext, onPrev }) => (
        <div className="flex flex-col gap-5 w-full">
          <div className="flex gap-15 mx-auto">
            <button className="default" onClick={onPrev} disabled={page <= 1}>
              Prev
            </button>
            <button
              className="default"
              onClick={onNext}
              disabled={!data.has_next_page}
            >
              Next
            </button>
          </div>
          <TopItemList
            ranked
            separators
            data={data}
            className="w-11/12 sm:w-[600px]"
            type="track"
          />
          <div className="flex gap-15 mx-auto">
            <button className="default" onClick={onPrev} disabled={page === 0}>
              Prev
            </button>
            <button
              className="default"
              onClick={onNext}
              disabled={!data.has_next_page}
            >
              Next
            </button>
          </div>
        </div>
      )}
    />
  );
}
