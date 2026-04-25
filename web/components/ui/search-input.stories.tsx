import { useState } from "react";
import type { Meta, StoryObj } from "@storybook/nextjs-vite";

import { SearchInput } from "./search-input";

const meta: Meta<typeof SearchInput> = {
  title: "UI/SearchInput",
  component: SearchInput,
  parameters: {
    layout: "centered",
    docs: {
      description: {
        component:
          'Generic search primitive: icon-prefixed `<input type="search">` with an optional trailing clear button. Controlled — the consumer owns state and can debounce externally with `useDebouncedValue`.',
      },
    },
  },
  decorators: [
    (Story) => (
      <div className="w-[380px]">
        <Story />
      </div>
    ),
  ],
};

export default meta;
type Story = StoryObj<typeof SearchInput>;

function Controlled({
  initial = "",
  clearable = false,
}: {
  initial?: string;
  clearable?: boolean;
}) {
  const [value, setValue] = useState(initial);
  return (
    <SearchInput
      placeholder="Search courses…"
      value={value}
      onChange={(event) => setValue(event.target.value)}
      onClear={clearable ? () => setValue("") : undefined}
    />
  );
}

export const Empty: Story = {
  render: () => <Controlled />,
};

export const WithValue: Story = {
  render: () => <Controlled initial="algorithms" />,
};

export const Clearable: Story = {
  render: () => <Controlled initial="algorithms" clearable />,
};

export const Disabled: Story = {
  render: () => (
    <SearchInput
      placeholder="Search courses…"
      value="algorithms"
      onChange={() => {}}
      disabled
    />
  ),
};
