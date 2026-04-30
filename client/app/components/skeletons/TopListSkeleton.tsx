interface Props {
  numItems: number;
}

export default function TopListSkeleton({ numItems }: Props) {
  return (
    <div className="w-[300px]">
      {[...Array(numItems)].map(() => (
        <div className="flex items-center gap-2 mb-[4px]">
          <div className="w-[40px] h-[40px] bg-(--color-bg-tertiary) rounded"></div>
          <div>
            <div className="h-[14px] w-[150px] bg-(--color-bg-tertiary) rounded"></div>
            <div className="h-[12px] w-[60px] bg-(--color-bg-tertiary) rounded"></div>
          </div>
        </div>
      ))}
    </div>
  );
}
