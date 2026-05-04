import PinnedItem from "./PinnedItem";

export default function PinnedItemGrid() {
  return (
    <div className="flex flex-wrap justify-center max-w-[750px] gap-10">
      <PinnedItem />
      <PinnedItem />
      <PinnedItem />
      <PinnedItem />
    </div>
  );
}
