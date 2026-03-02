export default function LoadingIndicator({ label = "Loading" }: { label?: string }) {
  return (
    <div
      role="status"
      aria-label={label}
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
    >
      // change the size of the indicater and color
      <div className="h-12 w-12 animate-spin rounded-full border-4 border-white border-t-transparent" />
      <span className="sr-only">{label}</span>
    </div>
  );
}