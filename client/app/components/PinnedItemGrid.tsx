import PinnedItem from "./PinnedItem";

export default function PinnedItemGrid() {
  return (
    <div className="flex flex-wrap justify-center gap-8">
      <PinnedItem />
      <PinnedItem />
      <PinnedItem />
      <PinnedItem />
    </div>
  );
}
