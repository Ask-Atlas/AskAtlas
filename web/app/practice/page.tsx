"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";

interface StudyGuide {
  id: string;
  name: string;
  topic: string;
  questionCount: number;
}

const studyGuides: StudyGuide[] = [
  {
    id: "1",
    name: "World Geography",
    topic: "Geography",
    questionCount: 3,
  },
  {
    id: "2",
    name: "Basic Science",
    topic: "Science",
    questionCount: 2,
  },
  {
    id: "3",
    name: "World History",
    topic: "History",
    questionCount: 2,
  }
];


export default function PracticePage() {
  const [selectedGuide, setSelectedGuide] = useState<StudyGuide | null>(null);
  
  // If a guide is selected, show a different screen
  if (selectedGuide) {
    return (
      <div className="min-h-screen bg-black text-white">
        <div className="max-w-7xl mx-auto px-6 py-12">
          <h1 className="text-3xl font-bold mb-4">
            You selected: {selectedGuide.name}
          </h1>
          <p className="text-gray-400">Quiz interface coming next!</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-white">
      {/* Hero Section */}
      <div className="relative overflow-hidden border-b border-white/10">
        <div className="absolute inset-0 bg-gradient-to-br from-orange-500/5 via-transparent to-blue-500/5" />
        <div className="relative max-w-7xl mx-auto px-6 py-16">
          <Badge className="mb-4 bg-orange-500/10 text-orange-500 border-orange-500/20">
            Practice Mode
          </Badge>
          <h1 className="text-5xl font-bold mb-4">
            Practice <span className="text-orange-500">Questions</span>
          </h1>
          <p className="text-xl text-gray-400 max-w-2xl">
            Practice by topic, check your progress, and spend more time where you need reinforcement.
          </p>
        </div>
      </div>

      {/* Study Guide Selection */}
      <div className="max-w-7xl mx-auto px-6 py-12">
        <h2 className="text-2xl font-semibold mb-6">Select a Study Guide</h2>
        
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {studyGuides.map((guide) => (
            <button
                key={guide.id}
                onClick={() => setSelectedGuide(guide)}
                className="text-left p-6 bg-white/5 border border-white/10 rounded-xl hover:border-orange-500/50 hover:bg-white/[0.07] transition-all duration-200 cursor-pointer"
            >
              <div className="flex items-start justify-between mb-4">
                <Badge variant="outline" className="border-blue-500/50 text-blue-400">
                  {guide.topic}
                </Badge>
                <span className="text-sm text-gray-400">{guide.questionCount} questions</span>
              </div>
              
              <h3 className="text-xl font-semibold mb-2">
                {guide.name}
              </h3>
              
              <div className="flex items-center text-sm text-gray-400 mt-4">
                <span>Start Practice</span>
                <span className="ml-2">→</span>
              </div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}