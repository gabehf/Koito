interface Props {
  size: number;
  src: string;
  className?: string;
}

export default function ActivityGrid(props: Props) {
  const classes = "image rounded-(--border-radius) border " + props.className;
  return (
    <>
      <img
        src={props.src}
        className={classes}
        style={{ height: props.size, width: props.size }}
      />
    </>
  );
}
