import type { DashboardCommonDictionary } from "../types";

const enDashboardCommonDictionary: DashboardCommonDictionary = {
  sidebar: {
    brandName: "AskAtlas",
    brandTagline: "Your study workspace",
    groupLabel: "Navigation",
    items: {
      home: "Home",
      starred: "Starred",
      courses: "Courses",
      studyGuides: "Study Guides",
      resources: "Resources",
      starredAll: "All Starred",
      starredRecent: "Recently Starred",
      coursesBrowse: "Browse Courses",
      coursesMine: "My Courses",
      coursesSectionMine: "My Courses",
      guidesCreate: "Create New Guide",
      guidesMine: "My Study Guides",
      guidesRecent: "Recently Viewed",
      resourcesUpload: "Upload Resource",
      resourcesView: "View Resources",
      resourcesRecent: "Recently Viewed",
      // TODO: Replace these with dynamic samples from the backend
      samples: {
        machineLearningFundamentals: "Machine Learning Fundamentals",
        binaryTreesCheatSheet: "Binary Trees Cheat Sheet",
        neuralNetworksPaper: "Neural Networks Paper",
        introPsychology: "Introduction to Psychology",
        dataStructuresAlgorithms: "Data Structures & Algorithms",
        modernWebDevelopment: "Modern Web Development",
        midtermReviewPsychology: "Midterm Review - Psychology",
        algorithmComplexityNotes: "Algorithm Complexity Notes",
        databasesQuickReference: "Databases Quick Reference",
        cloudComputingNotes: "Cloud Computing Notes",
      },
    },
  },
  breadcrumb: {
    home: "Home",
    browseCourses: "Browse Courses",
    myCourses: "My Courses",
    studyGuides: "Study Guides",
    createStudyGuide: "Create Study Guide",
    myStudyGuides: "My Study Guides",
    resources: "Resources",
    uploadResource: "Upload Resource",
    starred: "Starred",
  },
  userMenu: {
    upgradeToPro: "Upgrade to Pro",
    account: "Account",
    billing: "Billing",
    notifications: "Notifications",
    logout: "Log out",
  },
};

export default enDashboardCommonDictionary;
