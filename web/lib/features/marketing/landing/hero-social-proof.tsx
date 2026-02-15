"use client";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { BookOpen } from "lucide-react";
import { FEATURED_STUDENTS } from "./fixtures";
import { useLandingCopy } from "./i18n/landing-copy-provider";

export function HeroSocialProof() {
  const copy = useLandingCopy();

  return (
    <div className="flex flex-wrap items-center gap-6 pt-4 text-base text-muted-foreground">
      <div className="flex items-center gap-4">
        <TooltipProvider delayDuration={100}>
          <div className="flex -space-x-4">
            {FEATURED_STUDENTS.map((student) => (
              <Tooltip key={student.id}>
                <TooltipTrigger asChild>
                  <Avatar className="h-12 w-12 border-2 border-background cursor-pointer hover:z-10 transition-transform hover:scale-110">
                    <AvatarImage src={student.avatar} alt={student.name} />
                    <AvatarFallback className="bg-primary/10 text-primary font-medium text-xs">
                      {student.name
                        .split(" ")
                        .map((n) => n[0])
                        .join("")}
                    </AvatarFallback>
                  </Avatar>
                </TooltipTrigger>
                <TooltipContent>
                  <p className="font-medium">{student.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {student.university}
                  </p>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </TooltipProvider>
        <div className="flex flex-col">
          <span className="text-lg font-bold text-foreground">
            {copy.socialProof.studentsValue}
          </span>
          <span className="text-sm">{copy.socialProof.studentsLabel}</span>
        </div>
      </div>

      <div className="hidden sm:block h-8 w-px bg-primary/20" />

      <div className="flex items-center gap-2">
        <BookOpen className="h-5 w-5 text-primary" />
        <div className="flex flex-col">
          <span className="text-lg font-bold text-foreground">
            {copy.socialProof.classesValue}
          </span>
          <span className="text-sm">{copy.socialProof.classesLabel}</span>
        </div>
      </div>
    </div>
  );
}
