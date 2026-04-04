export interface StudyGuideResource {
  id: string;
  title: string;
  url: string;
  type: "pdf" | "link" | "video";
}

export interface StudyGuideQuiz {
  id: string;
  title: string;
  questionCount: number;
}

export interface StudyGuide {
  id: string;
  title: string;
  description: string;
  course: string;
  createdBy: string;
  updatedAt: string;
  quizzes: StudyGuideQuiz[];
  resources: StudyGuideResource[];
}
