"use client";

import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { SkeletonGrid } from "@/components/ui/skeleton-grid";
import { listCourses, listMyEnrollments, listSchools } from "@/lib/api";
import type {
  ListCoursesQuery,
  ListCoursesResponse,
  SchoolResponse,
} from "@/lib/api/types";
import { CourseCard } from "@/lib/features/dashboard/courses/course-card";
import { CourseSearchBar } from "@/lib/features/dashboard/courses/course-search-bar";
import { toast } from "@/lib/features/shared/toast/toast";

const GRID_CLASSES = "grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3";

export default function CoursesPage() {
  const [query, setQuery] = useState<ListCoursesQuery>({});
  const [data, setData] = useState<ListCoursesResponse | null>(null);
  const [schools, setSchools] = useState<SchoolResponse[]>([]);
  const [joinedCourseIds, setJoinedCourseIds] = useState<Set<string>>(
    () => new Set(),
  );
  const [isLoading, setIsLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    Promise.allSettled([
      listSchools({ page_limit: 100 }),
      listMyEnrollments(),
    ]).then(([schoolsRes, enrollmentsRes]) => {
      if (cancelled) return;
      if (schoolsRes.status === "fulfilled") {
        setSchools(schoolsRes.value.schools);
      }
      if (enrollmentsRes.status === "fulfilled") {
        setJoinedCourseIds(
          new Set(enrollmentsRes.value.enrollments.map((e) => e.course.id)),
        );
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    let cancelled = false;
    setIsLoading(true);
    setError(null);
    listCourses(query)
      .then((res) => {
        if (cancelled) return;
        setData(res);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : "Failed to load courses");
      })
      .finally(() => {
        if (cancelled) return;
        setIsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [query]);

  async function handleLoadMore() {
    if (!data?.next_cursor || isLoadingMore) return;
    setIsLoadingMore(true);
    try {
      const next = await listCourses({
        ...query,
        cursor: data.next_cursor,
      });
      setData((prev) =>
        prev
          ? {
              courses: [...prev.courses, ...next.courses],
              next_cursor: next.next_cursor,
              has_more: next.has_more,
            }
          : next,
      );
    } catch (err) {
      toast.error(err);
    } finally {
      setIsLoadingMore(false);
    }
  }

  const courses = data?.courses ?? [];
  const showEmptyState =
    !isLoading && !error && data !== null && courses.length === 0;

  return (
    <section className="space-y-6">
      <header className="space-y-1.5">
        <h1 className="text-2xl font-semibold tracking-tight">
          Browse courses
        </h1>
        <p className="text-muted-foreground text-sm">
          Discover courses across schools, follow them, and start a study guide.
        </p>
      </header>

      <CourseSearchBar value={query} onChange={setQuery} schools={schools} />

      {error ? (
        <EmptyState
          title="Couldn't load courses"
          body={error}
          action={
            <Button variant="outline" onClick={() => setQuery({ ...query })}>
              Try again
            </Button>
          }
        />
      ) : isLoading && data === null ? (
        <SkeletonGrid count={9} className={GRID_CLASSES} />
      ) : showEmptyState ? (
        <EmptyState
          title="No courses match"
          body="Try a different search or clear the filters."
          action={
            <Button variant="outline" onClick={() => setQuery({})}>
              Clear filters
            </Button>
          }
        />
      ) : (
        <>
          <p aria-live="polite" className="text-muted-foreground text-sm">
            {courses.length} course{courses.length === 1 ? "" : "s"}
            {query.q ? ` matching "${query.q}"` : ""}
          </p>

          <div className={GRID_CLASSES}>
            {courses.map((course) => (
              <CourseCard
                key={course.id}
                course={course}
                variant="tile"
                rightSlot={
                  joinedCourseIds.has(course.id) ? (
                    <Badge variant="secondary">Joined</Badge>
                  ) : undefined
                }
              />
            ))}
          </div>

          {data?.has_more ? (
            <div className="flex justify-center pt-2">
              <Button
                variant="outline"
                onClick={handleLoadMore}
                disabled={isLoadingMore}
              >
                {isLoadingMore ? (
                  <>
                    <Loader2 className="size-4 animate-spin" aria-hidden />
                    Loading…
                  </>
                ) : (
                  "Load more"
                )}
              </Button>
            </div>
          ) : null}
        </>
      )}
    </section>
  );
}
