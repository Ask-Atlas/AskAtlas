export interface DashboardCommonDictionary {
  sidebar: {
    brandName: string;
    brandTagline: string;
    groupLabel: string;
    items: {
      home: string;
      starred: string;
      courses: string;
      studyGuides: string;
      resources: string;
      starredAll: string;
      starredRecent: string;
      coursesBrowse: string;
      coursesMine: string;
      coursesSectionMine: string;
      guidesCreate: string;
      guidesMine: string;
      guidesRecent: string;
      resourcesUpload: string;
      resourcesView: string;
      resourcesRecent: string;
      // TODO: Replace these with dynamic samples from the backend
      samples: {
        machineLearningFundamentals: string;
        binaryTreesCheatSheet: string;
        neuralNetworksPaper: string;
        introPsychology: string;
        dataStructuresAlgorithms: string;
        modernWebDevelopment: string;
        midtermReviewPsychology: string;
        algorithmComplexityNotes: string;
        databasesQuickReference: string;
        cloudComputingNotes: string;
      };
    };
  };
  breadcrumb: {
    home: string;
    browseCourses: string;
    myCourses: string;
    studyGuides: string;
    createStudyGuide: string;
    myStudyGuides: string;
    resources: string;
    uploadResource: string;
    starred: string;
  };
  userMenu: {
    upgradeToPro: string;
    account: string;
    billing: string;
    notifications: string;
    logout: string;
  };
}
