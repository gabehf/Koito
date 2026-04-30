import React, { useState } from "react";

type Props = {
  children: React.ReactNode;
  onClick: () => void;
  loading?: boolean;
  disabled?: boolean;
  confirm?: boolean;
};

export function AsyncButton(props: Props) {
  const [awaitingConfirm, setAwaitingConfirm] = useState(false);

  const handleClick = () => {
    if (props.confirm) {
      if (!awaitingConfirm) {
        setAwaitingConfirm(true);
        setTimeout(() => setAwaitingConfirm(false), 3000);
        return;
      }
      setAwaitingConfirm(false);
    }

    props.onClick();
  };

  return (
    <button
      onClick={handleClick}
      disabled={props.loading || props.disabled}
      className={`relative px-5 py-2 rounded-md large-button flex disabled:opacity-50 items-center`}
    >
      <span className={props.loading ? "invisible" : "visible"}>
        {awaitingConfirm ? "Are you sure?" : props.children}
      </span>
      {props.loading && (
        <span className="absolute inset-0 flex items-center justify-center">
          <span className="animate-spin h-4 w-4 border-2 border-white border-t-transparent rounded-full"></span>
        </span>
      )}
    </button>
  );
}
