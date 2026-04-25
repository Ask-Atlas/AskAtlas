import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import type { ListCoursesQuery, SchoolResponse } from "@/lib/api/types";

import { CourseSearchBar } from "./course-search-bar";

const SCHOOLS: SchoolResponse[] = [
  {
    id: "s_preview_1",
    name: "Washington State University",
    acronym: "WSU",
    city: "Pullman",
    state: "WA",
    country: "US",
    created_at: "2026-04-20T10:00:00Z",
  },
  {
    id: "s_preview_2",
    name: "University of Washington",
    acronym: "UW",
    city: "Seattle",
    state: "WA",
    country: "US",
    created_at: "2026-04-20T10:00:00Z",
  },
];

const meta: Meta<typeof CourseSearchBar> = {
  title: "Dashboard/CourseSearchBar",
  component: CourseSearchBar,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          "Search + filter row for the course catalog and onboarding flow. Search input debounces by 250ms; school + department selects emit immediately. Built on the generic SearchInput primitive and useDebouncedValue hook, so the same controls compose into other surfaces (file picker, study-guide search) later.",
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[720px] max-w-full">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof CourseSearchBar>;

function Controlled({
  initial = {},
  departments,
}: {
  initial?: ListCoursesQuery;
  departments?: readonly string[];
}) {
  const [value, setValue] = useState<ListCoursesQuery>(initial);
  return (
    <div className="space-y-2">
      <CourseSearchBar
        value={value}
        onChange={setValue}
        schools={SCHOOLS}
        departments={departments}
      />
      <pre className="text-muted-foreground bg-muted/40 rounded-md border p-2 text-xs">
        {JSON.stringify(value, null, 2)}
      </pre>
    </div>
  );
}

export const Default: Story = {
  render: () => <Controlled />,
};

export const WithDepartments: Story = {
  render: () => <Controlled departments={["CPTS", "MATH", "PHYS", "ENGL"]} />,
};

export const Prefilled: Story = {
  render: () => (
    <Controlled
      initial={{ q: "algorithms", school_id: "s_preview_1" }}
      departments={["CPTS", "MATH"]}
    />
  ),
};
