interface Props {
  size: number;
  src: string;
  alt: string;
  className?: string;
  srcset?: string;
  sizes?: string;
  lazy?: boolean;
}

export default function ActivityGrid(props: Props) {
  const classes = "image rounded-(--border-radius) border " + props.className;
  return (
    <>
      <img
        src={props.src}
        className={classes}
        alt={props.alt}
        style={{ height: props.size, width: props.size }}
        loading={props.lazy ? "lazy" : undefined}
        srcSet={props.srcset}
        sizes={props.sizes}
      />
    </>
  );
}
