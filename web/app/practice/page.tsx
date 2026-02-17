import { Badge } from "@/components/ui/badge";

export default function PracticePage() {
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

      {/* Placeholder for content */}
      <div className="max-w-7xl mx-auto px-6 py-12">
        <p className="text-gray-400">Study guide selection coming next...</p>
      </div>
    </div>
  );
}