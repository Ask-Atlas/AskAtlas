"use client";

import { useEffect, useRef, useState } from "react";

import { SearchInput } from "@/components/ui/search-input";
import { useDebouncedValue } from "@/hooks/use-debounced-value";
import type { ListCoursesQuery, SchoolResponse } from "@/lib/api/types";
import { cn } from "@/lib/utils";

const ALL = "__all__";
const DEBOUNCE_MS = 250;

export interface CourseSearchBarProps {
  value: ListCoursesQuery;
  onChange: (next: ListCoursesQuery) => void;
  schools: readonly SchoolResponse[];
  departments?: readonly string[];
  className?: string;
}

export function CourseSearchBar({
  value,
  onChange,
  schools,
  departments,
  className,
}: CourseSearchBarProps) {
  const [q, setQ] = useState(value.q ?? "");
  const debouncedQ = useDebouncedValue(q, DEBOUNCE_MS);

  // Refs let the debounce-emit effect depend only on `debouncedQ`;
  // including `value`/`onChange` in deps would loop every render.
  const onChangeRef = useRef(onChange);
  const valueRef = useRef(value);
  useEffect(() => {
    onChangeRef.current = onChange;
    valueRef.current = value;
  });

  useEffect(() => {
    const trimmed = debouncedQ.trim();
    const nextQ = trimmed === "" ? undefined : trimmed;
    if (nextQ === (valueRef.current.q ?? undefined)) return;
    onChangeRef.current({ ...valueRef.current, q: nextQ });
  }, [debouncedQ]);

  function handleSchoolChange(next: string) {
    onChange({
      ...value,
      school_id: next === ALL ? undefined : next,
    });
  }

  function handleDepartmentChange(next: string) {
    onChange({
      ...value,
      department: next === ALL ? undefined : next,
    });
  }

  const filterClasses =
    "border-input bg-transparent text-foreground focus-visible:border-ring focus-visible:ring-ring/50 dark:bg-input/30 h-9 rounded-md border px-3 py-1 text-sm shadow-xs outline-none transition-[color,box-shadow] focus-visible:ring-[3px] disabled:cursor-not-allowed disabled:opacity-50";

  return (
    <div
      className={cn(
        "flex flex-col gap-2 sm:flex-row sm:items-center",
        className,
      )}
    >
      <SearchInput
        placeholder="Search courses"
        value={q}
        onChange={(event) => setQ(event.target.value)}
        onClear={q === "" ? undefined : () => setQ("")}
        aria-label="Search courses"
      />
      <select
        aria-label="Filter by school"
        value={value.school_id ?? ALL}
        onChange={(event) => handleSchoolChange(event.target.value)}
        className={cn(filterClasses, "sm:w-56")}
      >
        <option value={ALL}>All schools</option>
        {schools.map((school) => (
          <option key={school.id} value={school.id}>
            {school.name}
          </option>
        ))}
      </select>
      {departments && departments.length > 0 ? (
        <select
          aria-label="Filter by department"
          value={value.department ?? ALL}
          onChange={(event) => handleDepartmentChange(event.target.value)}
          className={cn(filterClasses, "sm:w-44")}
        >
          <option value={ALL}>All departments</option>
          {departments.map((dept) => (
            <option key={dept} value={dept}>
              {dept}
            </option>
          ))}
        </select>
      ) : null}
    </div>
  );
}
