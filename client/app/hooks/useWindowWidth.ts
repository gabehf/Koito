import { useState, useEffect } from "react";

export default function useWindowWidth() {
  const [width, setWidth] = useState(window.innerWidth);

  useEffect(() => {
    const handleResize = () => setWidth(window.innerWidth);
    window.addEventListener("resize", handleResize);

    // Clean up the listener when the component unmounts
    return () => window.removeEventListener("resize", handleResize);
  }, []);

  return width;
}
