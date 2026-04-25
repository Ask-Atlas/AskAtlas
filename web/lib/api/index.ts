/**
 * Barrel export for the typed API surface.
 *
 * Prefer:       import { listFiles, type FileResponse } from "@/lib/api";
 * Also fine:    import { listFiles } from "@/lib/api/actions/files";
 *
 * Both forms resolve to the same Server Action -- the barrel is for
 * ergonomics at the callsite and tree-shakes the same way.
 */

// Errors
export { ApiError, unwrap, unwrapVoid } from "./errors";

// Named schema aliases
export type * from "./types";

// Actions: one re-export block per domain so IDE import suggestions
// stay grouped by feature instead of drowning in a 50+ line list.

// --- Files ---
export {
  createFile,
  createGrant,
  deleteFile,
  getFile,
  listFiles,
  recordFileView,
  revokeGrant,
  toggleFileFavorite,
  updateFile,
} from "./actions/files";

// --- Schools ---
export { getSchool, listSchools } from "./actions/schools";

// --- Courses ---
export {
  checkMembership,
  createStudyGuideForCourse,
  getCourse,
  joinSection,
  leaveSection,
  listCourseSections,
  listCourseStudyGuides,
  listCourses,
  listSectionMembers,
} from "./actions/courses";

// --- Study Guides ---
export {
  attachFile,
  attachResource,
  castStudyGuideVote,
  createStudyGuideGrant,
  deleteStudyGuide,
  detachFile,
  detachResource,
  getStudyGuide,
  listStudyGuideGrants,
  recommendStudyGuide,
  removeStudyGuideRecommendation,
  removeStudyGuideVote,
  revokeStudyGuideGrant,
  updateStudyGuide,
} from "./actions/study-guides";

// --- Quizzes ---
export {
  addQuizQuestion,
  createQuiz,
  deleteQuiz,
  deleteQuizQuestion,
  getQuiz,
  listQuizzes,
  replaceQuizQuestion,
  updateQuiz,
} from "./actions/quizzes";

// --- Practice ---
export {
  abandonPracticeSession,
  completePracticeSession,
  getPracticeSession,
  listPracticeSessions,
  startPracticeSession,
  submitPracticeAnswer,
} from "./actions/practice";

// --- Me ---
export {
  listDashboard,
  listFavorites,
  listMyEnrollments,
  listMyStudyGuides,
  listRecents,
  toggleCourseFavorite,
  toggleStudyGuideFavorite,
} from "./actions/me";
