import TopItemList from "~/components/TopItemList";
import ChartLayout from "./ChartLayout";
import { useLoaderData, type LoaderFunctionArgs } from "react-router";
import { type Album, type PaginatedResponse, type Ranked } from "api/api";

export async function clientLoader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url);
  const page = url.searchParams.get("page") || "0";
  url.searchParams.set("page", page);

  const res = await fetch(
    `/apis/web/v1/top-artists?${url.searchParams.toString()}`
  );
  if (!res.ok) {
    throw new Response("Failed to load top artists", { status: 500 });
  }

  const top_artists: PaginatedResponse<Album> = await res.json();
  return { top_artists };
}

export default function Artist() {
  const { top_artists: initialData } = useLoaderData<{
    top_artists: PaginatedResponse<Ranked<Album>>;
  }>();

  return (
    <ChartLayout
      title="Top Artists"
      initialData={initialData}
      endpoint="chart/top-artists"
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
            type="artist"
          />
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
        </div>
      )}
    />
  );
}
