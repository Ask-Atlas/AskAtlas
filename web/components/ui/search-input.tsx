import * as React from "react";
import { Search, X } from "lucide-react";

import { cn } from "@/lib/utils";

export interface SearchInputProps extends Omit<
  React.ComponentProps<"input">,
  "type"
> {
  containerClassName?: string;
  onClear?: () => void;
  clearLabel?: string;
}

function SearchInput({
  className,
  containerClassName,
  onClear,
  clearLabel = "Clear search",
  value,
  disabled,
  ...props
}: SearchInputProps) {
  const hasValue = typeof value === "string" ? value.length > 0 : value != null;
  const showClear = onClear != null && hasValue && !disabled;

  return (
    <div
      data-slot="search-input"
      role="group"
      data-disabled={disabled ? "true" : undefined}
      className={cn(
        "border-input dark:bg-input/30 relative flex h-9 w-full min-w-0 items-center gap-2 rounded-md border bg-transparent pr-1.5 pl-3 shadow-xs transition-[color,box-shadow]",
        "has-[input:focus-visible]:border-ring has-[input:focus-visible]:ring-ring/50 has-[input:focus-visible]:ring-[3px]",
        "has-[input[aria-invalid=true]]:border-destructive has-[input[aria-invalid=true]]:ring-destructive/20 dark:has-[input[aria-invalid=true]]:ring-destructive/40",
        "data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50",
        containerClassName,
      )}
    >
      <Search
        aria-hidden
        className="text-muted-foreground pointer-events-none size-4 shrink-0"
      />
      {/* type="text" (not "search") avoids a duplicate native webkit clear "x" alongside our own. */}
      <input
        type="text"
        role="searchbox"
        value={value}
        disabled={disabled}
        data-slot="input"
        className={cn(
          "placeholder:text-muted-foreground selection:bg-primary selection:text-primary-foreground text-foreground h-full min-w-0 flex-1 bg-transparent text-base outline-none disabled:cursor-not-allowed md:text-sm",
          className,
        )}
        {...props}
      />
      {showClear ? (
        <button
          type="button"
          aria-label={clearLabel}
          onClick={onClear}
          className="text-muted-foreground hover:text-foreground focus-visible:ring-ring/50 inline-flex size-6 shrink-0 items-center justify-center rounded focus-visible:ring-[3px] focus-visible:outline-none"
        >
          <X aria-hidden className="size-3.5" />
        </button>
      ) : null}
    </div>
  );
}

export { SearchInput };
