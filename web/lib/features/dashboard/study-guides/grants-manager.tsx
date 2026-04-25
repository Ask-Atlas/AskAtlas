"use client";

import { Loader2, X } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import type {
  EnrollmentResponse,
  ListMyEnrollmentsResponse,
  ListStudyGuideGrantsResponse,
  StudyGuideCreateGrantRequest,
  StudyGuideGrantResponse,
  StudyGuideRevokeGrantRequest,
} from "@/lib/api/types";
import { toast } from "@/lib/features/shared/toast/toast";

const SEARCH_DEBOUNCE_MS = 200;

type GranteeType = StudyGuideRevokeGrantRequest["grantee_type"];
type Permission = StudyGuideRevokeGrantRequest["permission"];

function isGranteeType(value: string): value is GranteeType {
  return value === "course" || value === "user";
}

function isPermission(value: string): value is Permission {
  return value === "view" || value === "share" || value === "delete";
}

/**
 * Server actions are injected so this client component stays free of
 * `"use server"` imports. That keeps Storybook (Vite) from following
 * the chain into `@clerk/nextjs/server`, and matches the DI pattern
 * used by every other feature component in this codebase.
 */
export interface GrantsManagerActions {
  listGrants: (studyGuideId: string) => Promise<ListStudyGuideGrantsResponse>;
  listEnrollments: () => Promise<ListMyEnrollmentsResponse>;
  createGrant: (
    studyGuideId: string,
    body: StudyGuideCreateGrantRequest,
  ) => Promise<StudyGuideGrantResponse>;
  revokeGrant: (
    studyGuideId: string,
    body: StudyGuideRevokeGrantRequest,
  ) => Promise<void>;
}

export interface GrantsManagerProps {
  studyGuideId: string;
  actions: GrantsManagerActions;
  /** Optional callback invoked whenever the grant count changes. */
  onGrantCountChange?: (count: number) => void;
}

/**
 * Inline grants editor -- lists current non-creator grants on a study
 * guide and lets the caller add more. Loaded only when mounted (the
 * form keeps it unmounted in create mode, so the network fetch only
 * fires on the first popover open in edit mode).
 *
 * Updates are optimistic: we mutate the local `grants` state first,
 * fire the server action, and revert on failure. A toast surfaces the
 * error message for the user.
 *
 * People search is deferred -- typing `@<query>` shows a coming-soon
 * placeholder (see TODO below). For now we search only over the
 * caller's own enrollments (`GET /me/courses`).
 */
export function GrantsManager({
  studyGuideId,
  actions,
  onGrantCountChange,
}: GrantsManagerProps) {
  const [grants, setGrants] = useState<StudyGuideGrantResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [enrollments, setEnrollments] = useState<EnrollmentResponse[]>([]);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  // Course IDs with an in-flight POST. Gates duplicate clicks even
  // before the optimistic state has flushed.
  const inFlightAdds = useRef<Set<string>>(new Set());

  // Initial load -- grants + enrollments in parallel so the popover is
  // interactive as soon as the data lands.
  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [grantsRes, enrollmentsRes] = await Promise.all([
          actions.listGrants(studyGuideId),
          actions.listEnrollments(),
        ]);
        if (cancelled) return;
        setGrants(grantsRes.grants);
        setEnrollments(enrollmentsRes.enrollments);
      } catch (error: unknown) {
        if (!cancelled) toast.error(error);
      } finally {
        if (!cancelled) setIsLoading(false);
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [studyGuideId, actions]);

  // Report grant count changes back to the caller so the visibility
  // chip can re-render with the new "Shared · N" label.
  useEffect(() => {
    onGrantCountChange?.(grants.length);
  }, [grants.length, onGrantCountChange]);

  // Debounce the search query.
  useEffect(() => {
    const handle = setTimeout(
      () => setDebouncedQuery(query.trim()),
      SEARCH_DEBOUNCE_MS,
    );
    return () => clearTimeout(handle);
  }, [query]);

  const grantedCourseIds = useMemo(
    () =>
      new Set(
        grants
          .filter((grant) => grant.grantee_type === "course")
          .map((grant) => grant.grantee_id),
      ),
    [grants],
  );

  const filteredCourses = useMemo(() => {
    if (debouncedQuery === "" || debouncedQuery.startsWith("@")) return [];
    const needle = debouncedQuery.toLowerCase();
    // Dedupe by course.id -- a user can be enrolled in multiple
    // sections of the same course and we only grant at the course
    // level.
    const seen = new Set<string>();
    return enrollments
      .map((enrollment) => enrollment.course)
      .filter((course) => {
        if (seen.has(course.id) || grantedCourseIds.has(course.id))
          return false;
        const hay =
          `${course.department} ${course.number} ${course.title}`.toLowerCase();
        if (!hay.includes(needle)) return false;
        seen.add(course.id);
        return true;
      });
  }, [debouncedQuery, enrollments, grantedCourseIds]);

  const addCourseGrant = useCallback(
    async (courseId: string) => {
      if (inFlightAdds.current.has(courseId)) return;
      inFlightAdds.current.add(courseId);
      // Optimistic insert with a temporary id so React keys stay
      // stable; real id replaces it when the POST resolves.
      const temp: StudyGuideGrantResponse = {
        id: `optimistic-${courseId}`,
        study_guide_id: studyGuideId,
        grantee_type: "course",
        grantee_id: courseId,
        permission: "view",
        granted_by: "",
        created_at: new Date().toISOString(),
      };
      setGrants((prev) => [...prev, temp]);
      setQuery("");
      try {
        const created = await actions.createGrant(studyGuideId, {
          grantee_type: "course",
          grantee_id: courseId,
          permission: "view",
        });
        setGrants((prev) =>
          prev.map((grant) => (grant.id === temp.id ? created : grant)),
        );
      } catch (error: unknown) {
        setGrants((prev) => prev.filter((grant) => grant.id !== temp.id));
        toast.error(error);
      } finally {
        inFlightAdds.current.delete(courseId);
      }
    },
    [studyGuideId, actions],
  );

  const removeGrant = useCallback(
    async (grant: StudyGuideGrantResponse) => {
      if (
        !isGranteeType(grant.grantee_type) ||
        !isPermission(grant.permission)
      ) {
        toast.error(new Error("Cannot revoke grant: unrecognized type"));
        return;
      }
      // Functional setState captures the current array as `snapshot`
      // so a concurrent add/remove can't roll back to a stale closure.
      let snapshot: StudyGuideGrantResponse[] = [];
      setGrants((prev) => {
        snapshot = prev;
        return prev.filter((existing) => existing.id !== grant.id);
      });
      try {
        await actions.revokeGrant(studyGuideId, {
          grantee_type: grant.grantee_type,
          grantee_id: grant.grantee_id,
          permission: grant.permission,
        });
      } catch (error: unknown) {
        setGrants(snapshot);
        toast.error(error);
      }
    },
    [studyGuideId, actions],
  );

  const courseGrants = grants.filter(
    (grant) => grant.grantee_type === "course",
  );
  const userGrants = grants.filter((grant) => grant.grantee_type === "user");
  const showPeoplePlaceholder = debouncedQuery.startsWith("@");

  return (
    <div className="space-y-3">
      <div className="space-y-1">
        <p className="text-foreground text-xs font-medium">Share with</p>
        <Input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search your courses"
          aria-label="Search courses or people"
          className="h-8 text-xs"
        />
      </div>

      {isLoading ? (
        <div className="text-muted-foreground flex items-center gap-2 text-xs">
          <Loader2 className="size-3 animate-spin" aria-hidden />
          Loading grants…
        </div>
      ) : null}

      {!isLoading && debouncedQuery !== "" ? (
        showPeoplePlaceholder ? (
          // TODO(ASK-212-followup): wire up user search once a
          // `/users/search` endpoint exists. The `@` prefix signals
          // "search people" today but we stub it out.
          <p className="text-muted-foreground rounded-md border border-dashed px-2 py-1.5 text-xs">
            People search coming soon.
          </p>
        ) : filteredCourses.length === 0 ? (
          <p className="text-muted-foreground text-xs">No matching courses.</p>
        ) : (
          <ul className="space-y-1">
            {filteredCourses.map((course) => (
              <li key={course.id}>
                <button
                  type="button"
                  onClick={() => void addCourseGrant(course.id)}
                  className="hover:bg-muted focus-visible:ring-ring w-full rounded-md px-2 py-1.5 text-left text-xs transition-colors focus-visible:outline-none focus-visible:ring-1"
                >
                  <span className="font-medium">
                    {course.department} {course.number}
                  </span>
                  <span className="text-muted-foreground ml-1">
                    · {course.title}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        )
      ) : null}

      {!isLoading ? (
        <div className="space-y-2">
          <GrantGroup
            title="Courses"
            grants={courseGrants}
            enrollments={enrollments}
            onRemove={removeGrant}
          />
          <GrantGroup
            title="People"
            grants={userGrants}
            enrollments={enrollments}
            onRemove={removeGrant}
          />
          {courseGrants.length === 0 && userGrants.length === 0 ? (
            <p className="text-muted-foreground text-xs">
              No shares yet. Search above to add a course.
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

interface GrantGroupProps {
  title: string;
  grants: StudyGuideGrantResponse[];
  enrollments: EnrollmentResponse[];
  onRemove: (grant: StudyGuideGrantResponse) => void;
}

function GrantGroup({ title, grants, enrollments, onRemove }: GrantGroupProps) {
  if (grants.length === 0) return null;
  return (
    <div className="space-y-1">
      <p className="text-muted-foreground text-[11px] font-medium uppercase tracking-wide">
        {title}
      </p>
      <div className="flex flex-wrap gap-1">
        {grants.map((grant) => (
          <GrantChip
            key={grant.id}
            grant={grant}
            enrollments={enrollments}
            onRemove={onRemove}
          />
        ))}
      </div>
    </div>
  );
}

interface GrantChipProps {
  grant: StudyGuideGrantResponse;
  enrollments: EnrollmentResponse[];
  onRemove: (grant: StudyGuideGrantResponse) => void;
}

function GrantChip({ grant, enrollments, onRemove }: GrantChipProps) {
  const label =
    grant.grantee_type === "course"
      ? formatCourseLabel(grant.grantee_id, enrollments)
      : "User";
  return (
    <Badge
      variant="secondary"
      className="gap-1 pl-2 pr-1 text-xs font-normal"
      data-testid={`grant-chip-${grant.grantee_id}`}
    >
      {label}
      <button
        type="button"
        aria-label={`Remove ${label}`}
        onClick={() => onRemove(grant)}
        className="hover:bg-muted-foreground/20 focus-visible:ring-ring -mr-0.5 inline-flex size-4 items-center justify-center rounded-full focus-visible:outline-none focus-visible:ring-1"
      >
        <X className="size-3" aria-hidden />
      </button>
    </Badge>
  );
}

function formatCourseLabel(
  courseId: string,
  enrollments: EnrollmentResponse[],
): string {
  const match = enrollments.find(
    (enrollment) => enrollment.course.id === courseId,
  );
  // TODO(ASK-212-followup): the API should return course
  // department/number alongside each grant so we can label grants on
  // courses the viewer isn't enrolled in. For now we fall back to the
  // grantee_id (UUID) so duplicate "Course" chips remain distinguishable.
  if (!match) return `Course ${courseId.slice(0, 8)}`;
  return `${match.course.department} ${match.course.number}`;
}
