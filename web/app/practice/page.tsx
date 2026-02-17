"use client";

import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";

type QuestionType = "multiple-choice" | "true-false" | "freeform";

interface Question {
    id: string;
    type: QuestionType;
    question: string;
    options?: string[];
    correctAnswer: string | boolean;
    hint: string;
    feedback: {
        correct: string;
        incorrect: string;
    };
}

interface StudyGuide {
    id: string;
    name: string;
    topic: string;
    questionCount: number;
    questions: Question[];
}

const studyGuides: StudyGuide[] = [
    {
        id: "1",
        name: "World Geography",
        topic: "Geography",
        questionCount: 3,
        questions: [
            {
                id: "1",
                type: "multiple-choice",
                question: "What is the capital of France?",
                options: ["London", "Paris", "Berlin", "Madrid"],
                correctAnswer: "Paris",
                hint: "Think of the Eiffel Tower!",
                feedback: {
                    correct: "Excellent! Paris is indeed the capital and largest city of France.",
                    incorrect: "Not quite. Paris is the capital of France, famous for the Eiffel Tower and the Louvre."
                }
            },
            {
                id: "2",
                type: "freeform",
                question: "What is the longest river in the world?",
                correctAnswer: "Nile",
                hint: "It flows through northeastern Africa.",
                feedback: {
                    correct: "Perfect! The Nile River is approximately 6,650 km long.",
                    incorrect: "The correct answer is the Nile River, which flows through northeastern Africa."
                }
            },
            {
                id: "3",
                type: "true-false",
                question: "Mount Everest is located in Japan.",
                correctAnswer: false,
                hint: "It's in the Himalayas, on the border between Nepal and Tibet.",
                feedback: {
                    correct: "Correct! Mount Everest is on the Nepal-Tibet border, not in Japan.",
                    incorrect: "Actually, Mount Everest is located on the border between Nepal and Tibet, not in Japan."
                }
            }
        ]
    },
    {
        id: "2",
        name: "Basic Science",
        topic: "Science",
        questionCount: 2,
        questions: [
            {
                id: "4",
                type: "true-false",
                question: "The Earth is flat.",
                correctAnswer: false,
                hint: "Consider modern scientific evidence.",
                feedback: {
                    correct: "Correct! The Earth is an oblate spheroid.",
                    incorrect: "The Earth is round, confirmed by scientific evidence."
                }
            },
            {
                id: "5",
                type: "multiple-choice",
                question: "What is the chemical symbol for water?",
                options: ["H2O", "CO2", "O2", "NaCl"],
                correctAnswer: "H2O",
                hint: "It's made of hydrogen and oxygen.",
                feedback: {
                    correct: "Excellent! H2O represents two hydrogen and one oxygen.",
                    incorrect: "The correct answer is H2O."
                }
            }
        ]
    },
    {
        id: "3",
        name: "World History",
        topic: "History",
        questionCount: 2,
        questions: [
            {
                id: "6",
                type: "freeform",
                question: "What year did World War II end?",
                correctAnswer: "1945",
                hint: "It ended in the mid-1940s.",
                feedback: {
                    correct: "Perfect! World War II ended in 1945.",
                    incorrect: "World War II ended in 1945."
                }
            },
            {
                id: "7",
                type: "multiple-choice",
                question: "Who was the first President of the United States?",
                options: ["Thomas Jefferson", "George Washington", "John Adams", "Benjamin Franklin"],
                correctAnswer: "George Washington",
                hint: "He's on the one dollar bill.",
                feedback: {
                    correct: "Correct! George Washington served from 1789-1797.",
                    incorrect: "George Washington was the first U.S. President."
                }
            }
        ]
    }
];

export default function PracticePage() {
    const [selectedGuide, setSelectedGuide] = useState<StudyGuide | null>(null);
    const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
    const [userAnswer, setUserAnswer] = useState<string | null>(null);
    const [showFeedback, setShowFeedback] = useState(false);
    const [isCorrect, setIsCorrect] = useState<boolean | null>(null);
    const [showHint, setShowHint] = useState(false);

    // If a guide is selected, show the quiz
    if (selectedGuide) {
        const currentQuestion = selectedGuide.questions[currentQuestionIndex];

        const checkAnswer = () => {
            if (!currentQuestion || !userAnswer) return;

            const correct = userAnswer.toLowerCase().trim() === currentQuestion.correctAnswer.toString().toLowerCase().trim();
            setIsCorrect(correct);
            setShowFeedback(true);
        };

        const nextQuestion = () => {
            if (currentQuestionIndex < selectedGuide.questions.length - 1) {
                setCurrentQuestionIndex(prev => prev + 1);
                setUserAnswer(null);
                setShowHint(false);
                setShowFeedback(false);
                setIsCorrect(null);
            }
        };

        return (
            <div className="min-h-screen bg-black text-white">
                {/* Header with Back Button */}
                <div className="border-b border-white/10 bg-white/5">
                    <div className="max-w-7xl mx-auto px-6 py-4">
                        <div className="flex items-center justify-between">
                            <div className="flex items-center gap-4">
                                <Button
                                    onClick={() => {
                                        setSelectedGuide(null);
                                        setCurrentQuestionIndex(0);
                                        setUserAnswer(null);
                                        setShowHint(false);
                                        setShowFeedback(false);
                                    }}
                                    variant="ghost"
                                    size="sm"
                                    className="text-gray-400 hover:text-white"
                                >
                                    ← Back
                                </Button>
                                <div>
                                    <h2 className="font-semibold">{selectedGuide.name}</h2>
                                    <p className="text-sm text-gray-400">{selectedGuide.topic}</p>
                                </div>
                            </div>
                            <div>
                                <p className="text-sm text-gray-400">Progress</p>
                                <p className="text-lg font-semibold">
                                    {currentQuestionIndex + 1} / {selectedGuide.questions.length}
                                </p>
                            </div>
                        </div>
                    </div>
                </div>

                <div className="max-w-4xl mx-auto px-6 py-12">
                    {currentQuestion ? (
                        <div className="bg-white/5 border border-white/10 rounded-xl p-8">
                            <Badge
                                className={cn(
                                    "mb-6",
                                    currentQuestion.type === "multiple-choice" && "border-blue-500/50 text-blue-400",
                                    currentQuestion.type === "true-false" && "border-green-500/50 text-green-400",
                                    currentQuestion.type === "freeform" && "border-purple-500/50 text-purple-400"
                                )}
                                variant="outline"
                            >
                                {currentQuestion.type === "multiple-choice" && "Multiple Choice"}
                                {currentQuestion.type === "true-false" && "True / False"}
                                {currentQuestion.type === "freeform" && "Free Response"}
                            </Badge>

                            <h3 className="text-xl font-semibold mb-6">
                                {currentQuestion.question}
                            </h3>

                            {/* Multiple Choice */}
                            {currentQuestion.type === "multiple-choice" && currentQuestion.options && (
                                <div className="space-y-3 mb-6">
                                    {currentQuestion.options.map((option, index) => (
                                        <button
                                            key={index}
                                            onClick={() => setUserAnswer(option)}
                                            disabled={showFeedback}
                                            className={`w-full text-left p-4 rounded-xl border-2 transition-all duration-200 ${userAnswer === option
                                                    ? "border-orange-500 bg-orange-500/10"
                                                    : "border-white/10 hover:border-orange-500/50 hover:bg-orange-500/5"
                                                }`}
                                        >
                                            {option}
                                        </button>
                                    ))}
                                </div>
                            )}

                            {/* True/False */}
                            {currentQuestion.type === "true-false" && (
                                <div className="grid grid-cols-2 gap-4 mb-6">
                                    {["True", "False"].map((option) => (
                                        <button
                                            key={option}
                                            onClick={() => setUserAnswer(option)}
                                            disabled={showFeedback}
                                            className={`p-6 rounded-xl border-2 transition-all duration-200 font-semibold text-lg ${userAnswer === option
                                                    ? "border-orange-500 bg-orange-500/10"
                                                    : "border-white/10 hover:border-orange-500/50 hover:bg-orange-500/5"
                                                }`}
                                        >
                                            {option}
                                        </button>
                                    ))}
                                </div>
                            )}

                            {/* Freeform */}
                            {currentQuestion.type === "freeform" && (
                                <div className="mb-6">
                                    <Input
                                        type="text"
                                        placeholder="Type your answer here..."
                                        value={userAnswer || ""}
                                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUserAnswer(e.target.value)}
                                        disabled={showFeedback}
                                        className="w-full p-4 text-lg bg-white/5 border-2 border-white/10 rounded-xl focus:border-orange-500"
                                    />
                                </div>
                            )}

                            {!showFeedback && (
                                <div className="mb-6">
                                    <Button
                                        onClick={() => setShowHint(!showHint)}
                                        variant="outline"
                                        size="sm"
                                        className="border-yellow-500/30 text-yellow-500 hover:bg-yellow-500/10"
                                    >
                                        {showHint ? "Hide Hint" : "💡 Show Hint"}
                                    </Button>

                                    {showHint && (
                                        <div className="mt-4 p-4 bg-yellow-500/5 border border-yellow-500/20 rounded-lg">
                                            <p className="text-sm text-yellow-200">{currentQuestion.hint}</p>
                                        </div>
                                    )}
                                </div>
                            )}

                            {!showFeedback ? (
                                <Button
                                    onClick={checkAnswer}
                                    disabled={!userAnswer}
                                    className="w-full bg-orange-500 hover:bg-orange-600 text-white"
                                >
                                    Submit Answer
                                </Button>
                            ) : (
                                <>
                                    <div className={`mb-6 p-6 rounded-xl border-2 ${isCorrect
                                            ? "bg-green-500/10 border-green-500/50"
                                            : "bg-red-500/10 border-red-500/50"
                                        }`}>
                                        <div className="flex items-start gap-4">
                                            <div className="text-3xl">
                                                {isCorrect ? "🎉" : "📚"}
                                            </div>
                                            <div>
                                                <h3 className={`text-lg font-semibold mb-2 ${isCorrect ? "text-green-400" : "text-red-400"
                                                    }`}>
                                                    {isCorrect ? "Correct!" : "Not quite"}
                                                </h3>
                                                <p className="text-gray-300">
                                                    {isCorrect ? currentQuestion.feedback.correct : currentQuestion.feedback.incorrect}
                                                </p>
                                            </div>
                                        </div>
                                    </div>

                                    {currentQuestionIndex < selectedGuide.questions.length - 1 ? (
                                        <Button
                                            onClick={nextQuestion}
                                            className="w-full bg-orange-500 hover:bg-orange-600 text-white"
                                        >
                                            Next Question →
                                        </Button>
                                    ) : (
                                        <Button
                                            onClick={() => {
                                                setSelectedGuide(null);
                                                setCurrentQuestionIndex(0);
                                                setUserAnswer(null);
                                                setShowHint(false);
                                                setShowFeedback(false);
                                            }}
                                            className="w-full bg-orange-500 hover:bg-orange-600 text-white"
                                        >
                                            Finish Practice
                                        </Button>
                                    )}
                                </>
                            )}
                        </div>
                    ) : (
                        <p className="text-gray-400">No questions available yet!</p>
                    )}
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