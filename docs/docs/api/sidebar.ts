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
          id: "api/create-file",
          label: "Create a file reference and get a presigned upload URL",
          className: "api-method post",
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
        {
          type: "doc",
          id: "api/create-grant",
          label: "Grant a permission on a file",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/revoke-grant",
          label: "Revoke a permission on a file",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/list-schools",
          label: "List and search schools",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/get-school",
          label: "Get a single school by ID",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/list-courses",
          label: "List and search courses",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/get-course",
          label: "Get a course detail with embedded sections",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/get-study-guide",
          label: "Get a study guide detail",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/update-study-guide",
          label: "Update a study guide",
          className: "api-method patch",
        },
        {
          type: "doc",
          id: "api/delete-study-guide",
          label: "Soft-delete a study guide",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/attach-resource",
          label: "Attach an external resource to a study guide",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/detach-resource",
          label: "Detach a resource from a study guide",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/attach-file",
          label: "Attach a file to a study guide",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/detach-file",
          label: "Detach a file from a study guide",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/list-quizzes",
          label: "List quizzes attached to a study guide",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/create-quiz",
          label: "Create a quiz attached to a study guide",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/get-quiz",
          label: "Get a quiz with all questions and correct answers",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/update-quiz",
          label: "Update a quiz's metadata (title and/or description)",
          className: "api-method patch",
        },
        {
          type: "doc",
          id: "api/delete-quiz",
          label: "Soft-delete a quiz",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/add-quiz-question",
          label: "Add a question to an existing quiz",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/recommend-study-guide",
          label: "Recommend a study guide",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/remove-study-guide-recommendation",
          label: "Remove the authenticated user's recommendation on a study guide",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/cast-study-guide-vote",
          label: "Cast or change a vote on a study guide",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/remove-study-guide-vote",
          label: "Remove the authenticated user's vote on a study guide",
          className: "api-method delete",
        },
        {
          type: "doc",
          id: "api/list-study-guides",
          label: "List study guides for a course",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/create-study-guide",
          label: "Create a study guide for a course",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/list-section-members",
          label: "List the members of a course section",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/join-section",
          label: "Join a section as the authenticated user",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "api/list-my-enrollments",
          label: "List the authenticated user's section enrollments",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/check-membership",
          label: "Check the authenticated user's membership in a section",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "api/leave-section",
          label: "Leave a section as the authenticated user",
          className: "api-method delete",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
