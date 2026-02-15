import { cn } from "@/lib/utils";

interface SectionProps {
  children: React.ReactNode;
  className?: string;
  contentClassName?: string;
  id?: string;
  background?: React.ReactNode;
}

export function Section({
  children,
  className,
  contentClassName,
  id,
  background,
}: SectionProps) {
  return (
    <section id={id} className={cn("relative", className)}>
      {background && (
        <div className="absolute inset-0 overflow-hidden">{background}</div>
      )}
      <div
        className={cn(
          "relative mx-auto max-w-7xl px-4 py-16 lg:px-8 lg:py-24",
          contentClassName,
        )}
      >
        {children}
      </div>
    </section>
  );
}
