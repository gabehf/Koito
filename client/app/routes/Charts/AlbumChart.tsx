import TopItemList from "~/components/TopItemList";
import ChartLayout from "./ChartLayout";
import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import { type Album, type PaginatedResponse } from "api/api";

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = url.searchParams.get("page") || "0";
  url.searchParams.set("page", page);

  const res = await fetch(
    `/apis/web/v1/top-albums?${url.searchParams.toString()}`
  );
  if (!res.ok) {
    throw new Response("Failed to load top albums", { status: 500 });
  }

  const top_albums: PaginatedResponse<Album> = await res.json();
  return { top_albums };
}

export default function AlbumChart() {
  const { top_albums: initialData } = useLoaderData<{
    top_albums: PaginatedResponse<Album>;
  }>();

  return (
    <ChartLayout
      title="Top Albums"
      initialData={initialData}
      endpoint="chart/top-albums"
      render={({ data, page, onNext, onPrev }) => (
        <div className="flex flex-col gap-5">
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
            className="w-[400px] sm:w-[600px]"
            type="album"
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
