import { clsx } from "clsx";
import { useRef, useState, type ReactNode } from "react";

type Variant = "action" | "ghost";

type Props = {
  children: ReactNode;
  onActivate: () => void;
  dwell: boolean;
  dwellMs?: number;
  variant?: Variant;
  disabled?: boolean;
  pressed?: boolean;
  ariaLabel?: string;
  className?: string;
  testId?: string;
};

export function BigButton({
  children,
  onActivate,
  dwell,
  dwellMs = 1100,
  variant = "ghost",
  disabled = false,
  pressed,
  ariaLabel,
  className,
  testId,
}: Props) {
  const [progress, setProgress] = useState(0);
  const frame = useRef(0);
  const fired = useRef(false);

  function holdStart() {
    if (!dwell || disabled) return;
    fired.current = false;
    const started = performance.now();
    const step = (now: number) => {
      const ratio = Math.min(1, (now - started) / dwellMs);
      setProgress(ratio);
      if (ratio >= 1) {
        cancelAnimationFrame(frame.current);
        setProgress(0);
        fired.current = true;
        onActivate();
        return;
      }
      frame.current = requestAnimationFrame(step);
    };
    cancelAnimationFrame(frame.current);
    frame.current = requestAnimationFrame(step);
  }

  function holdEnd() {
    cancelAnimationFrame(frame.current);
    setProgress(0);
  }

  return (
    <button
      type="button"
      data-testid={testId}
      disabled={disabled}
      aria-label={ariaLabel}
      aria-pressed={pressed}
      onClick={() => {
        if (fired.current) {
          fired.current = false;
          return;
        }
        onActivate();
      }}
      onPointerDown={holdStart}
      onPointerUp={holdEnd}
      onPointerLeave={holdEnd}
      onPointerCancel={holdEnd}
      className={clsx(
        "relative isolate min-h-14 overflow-hidden rounded-xl border-2 px-4 py-2 text-center text-lg font-bold leading-tight",
        "disabled:cursor-not-allowed disabled:opacity-50",
        variant === "action"
          ? "border-action bg-action text-on-action"
          : "border-line bg-card text-ink",
        className,
      )}
    >
      <span className="relative z-10">{children}</span>
      {dwell && progress > 0 ? (
        <span className="absolute inset-x-0 bottom-0 z-20 h-1.5 bg-ink/20" aria-hidden>
          <span className="block h-full bg-action" style={{ width: `${Math.round(progress * 100)}%` }} />
        </span>
      ) : null}
    </button>
  );
}
