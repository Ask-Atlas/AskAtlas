from __future__ import annotations

from collections.abc import Mapping
from typing import TYPE_CHECKING, Any, TypeVar

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.dashboard_courses_section import DashboardCoursesSection
    from ..models.dashboard_files_section import DashboardFilesSection
    from ..models.dashboard_practice_section import DashboardPracticeSection
    from ..models.dashboard_study_guides_section import DashboardStudyGuidesSection


T = TypeVar("T", bound="DashboardResponse")


@_attrs_define
class DashboardResponse:
    """Response envelope for GET /api/me/dashboard. All four
    sections are always present (never null) so the frontend
    can render the home page from a single response. Each
    section's count fields are 0 and list fields are [] when
    the viewer has no data in that area.

        Attributes:
            courses (DashboardCoursesSection): Enrollment summary block. `current_term` is null when the
                viewer has no enrollments. `enrolled_count` is the number
                of courses the viewer is enrolled in for the resolved
                current term (NOT the lifetime enrollment total). `courses`
                is capped at 10.
            study_guides (DashboardStudyGuidesSection): Study-guide summary block. `created_count` excludes
                soft-deleted guides. `recent` is the 5 most recently
                updated guides the viewer created.
            practice (DashboardPracticeSection): Practice stats block. `sessions_completed` counts only
                sessions where `completed_at IS NOT NULL`.
                `total_questions_answered` is the sum of submitted answers
                across all completed sessions (from practice_answers, not
                the snapshot total on practice_sessions). `overall_accuracy`
                is the rounded percentage of correct/total across completed
                sessions; 0 when no sessions are completed (NULLIF prevents
                division by zero). `recent_sessions` is the 5 most recently
                completed sessions.
            files (DashboardFilesSection): File totals block. `total_count` and `total_size` exclude
                files in any deletion lifecycle and only count files with
                upload_status='complete'. `total_size` is bytes (int64).
                `recent` is the 5 most recently updated complete files.
    """

    courses: DashboardCoursesSection
    study_guides: DashboardStudyGuidesSection
    practice: DashboardPracticeSection
    files: DashboardFilesSection
    additional_properties: dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> dict[str, Any]:
        courses = self.courses.to_dict()

        study_guides = self.study_guides.to_dict()

        practice = self.practice.to_dict()

        files = self.files.to_dict()

        field_dict: dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "courses": courses,
                "study_guides": study_guides,
                "practice": practice,
                "files": files,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: type[T], src_dict: Mapping[str, Any]) -> T:
        from ..models.dashboard_courses_section import DashboardCoursesSection
        from ..models.dashboard_files_section import DashboardFilesSection
        from ..models.dashboard_practice_section import DashboardPracticeSection
        from ..models.dashboard_study_guides_section import DashboardStudyGuidesSection

        d = dict(src_dict)
        courses = DashboardCoursesSection.from_dict(d.pop("courses"))

        study_guides = DashboardStudyGuidesSection.from_dict(d.pop("study_guides"))

        practice = DashboardPracticeSection.from_dict(d.pop("practice"))

        files = DashboardFilesSection.from_dict(d.pop("files"))

        dashboard_response = cls(
            courses=courses,
            study_guides=study_guides,
            practice=practice,
            files=files,
        )

        dashboard_response.additional_properties = d
        return dashboard_response

    @property
    def additional_keys(self) -> list[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
