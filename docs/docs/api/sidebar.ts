import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "api/askatlas-api",
    },
    {
      type: "category",
      label: "UNTAGGED",
      items: [
        {
          type: "doc",
          id: "api/list-files",
          label: "List files for the current user",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/get-file",
          label: "Get a file by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/update-file",
          label: "Rename a file",
          className: "api-method patch",
        },
        {
          type: "doc",
          id: "api/delete-file",
          label: "Delete a file by ID",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
