import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { BookOpen, ExternalLink, FileText, Play, Video } from "lucide-react";
import Link from "next/link";
import type { StudyGuide, StudyGuideResource } from "./study-guide-view.types";

const MOCK_STUDY_GUIDE: StudyGuide = {
  id: "binary-trees-cheat-sheet",
  title: "Binary Trees Cheat Sheet",
  description:
    "A comprehensive overview of binary tree data structures, traversal algorithms, and common interview patterns.",
  course: "Data Structures & Algorithms",
  createdBy: "NathanielGainesWSU",
  updatedAt: "March 2026",
  quizzes: [
    { id: "q1", title: "Tree Traversal Quiz", questionCount: 10 },
    { id: "q2", title: "Balanced Trees Quiz", questionCount: 8 },
  ],
  resources: [
    {
      id: "r1",
      title: "Binary Trees - Visual Reference",
      url: "https://visualgo.net/en/bst",
      type: "link",
    },
    {
      id: "r2",
      title: "Lecture Slides - Week 7",
      url: "https://example.com/slides.pdf",
      type: "pdf",
    },
    {
      id: "r3",
      title: "Tree Algorithms Explained",
      url: "https://example.com/video",
      type: "video",
    },
  ],
};

function resourceIcon(type: StudyGuideResource["type"]) {
  if (type === "pdf") return <FileText className="h-4 w-4 shrink-0" />;
  if (type === "video") return <Video className="h-4 w-4 shrink-0" />;
  return <ExternalLink className="h-4 w-4 shrink-0" />;
}

export function StudyGuideView() {
  const guide = MOCK_STUDY_GUIDE;

  return (
    <section className="space-y-6">
      <header className="space-y-2">
        <div className="flex items-center gap-2">
          <Badge variant="secondary">{guide.course}</Badge>
        </div>
        <h1 className="text-2xl font-semibold tracking-tight">{guide.title}</h1>
        <p className="text-muted-foreground text-sm">{guide.description}</p>
        <p className="text-muted-foreground text-xs">
          Created by {guide.createdBy} · Updated {guide.updatedAt}
        </p>
      </header>

      <Separator />

      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <BookOpen className="text-muted-foreground h-4 w-4" />
          <h2 className="text-sm font-medium">Quizzes</h2>
        </div>
        {guide.quizzes.length === 0 ? (
          <div className="bg-muted/50 rounded-xl px-4 py-8 text-center">
            <p className="text-muted-foreground text-sm">No quizzes yet.</p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {guide.quizzes.map((quiz) => (
              <div
                key={quiz.id}
                className="bg-muted/50 flex items-center justify-between rounded-xl px-4 py-3"
              >
                <div>
                  <p className="text-sm font-medium">{quiz.title}</p>
                  <p className="text-muted-foreground text-xs">
                    {quiz.questionCount} questions
                  </p>
                </div>
                <Button size="sm" asChild>
                  <Link href="/practice">
                    <Play className="mr-1 h-3 w-3" />
                    Start
                  </Link>
                </Button>
              </div>
            ))}
          </div>
        )}
      </div>

      <Separator />

      <div className="space-y-3">
        <div className="flex items-center gap-2">
          <ExternalLink className="text-muted-foreground h-4 w-4" />
          <h2 className="text-sm font-medium">Referenced Resources</h2>
        </div>
        {guide.resources.length === 0 ? (
          <div className="bg-muted/50 rounded-xl px-4 py-8 text-center">
            <p className="text-muted-foreground text-sm">
              No resources linked.
            </p>
          </div>
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {guide.resources.map((resource) => (
              <a
                key={resource.id}
                href={resource.url}
                target="_blank"
                rel="noopener noreferrer"
                className="bg-muted/50 hover:bg-muted flex items-center gap-3 rounded-xl px-4 py-3 transition-colors"
              >
                {resourceIcon(resource.type)}
                <span className="text-sm font-medium">{resource.title}</span>
              </a>
            ))}
          </div>
        )}
      </div>
    </section>
  );
}
