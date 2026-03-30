import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "./card";
import { LucideIcon } from "lucide-react";

interface StatCardProps {
  title: string;
  value: string | number;
  icon: LucideIcon;
  description?: string;
  trend?: {
    value: number;
    isPositive: boolean;
  };
  secondaryValue?: string | number;
  secondaryLabel?: string;
}

export function StatCard({
  title,
  value,
  icon: Icon,
  description,
  trend,
  secondaryValue,
  secondaryLabel,
}: StatCardProps) {
  return (
    <Card className="bg-[#252525] border-[#3a3a3a] flex flex-col justify-center">
      <CardHeader className="flex flex-col items-center justify-center space-y-1 pb-0 pt-3">
        <Icon className="h-4 w-4 text-orange-400" />
        <CardTitle className="text-sm font-medium text-gray-300 text-center">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="flex flex-col items-center justify-center py-3">
        {secondaryValue !== undefined && secondaryLabel ? (
          <div className="flex items-center gap-6">
            <div className="flex flex-col items-center">
              <div className="text-2xl font-bold text-amber-400">
                {value}
              </div>
              {description && (
                <p className="text-xs text-gray-400 text-center mt-0.5">
                  {description}
                </p>
              )}
            </div>
            <div className="h-10 w-px bg-zinc-700"></div>
            <div className="flex flex-col items-center">
              <div className="text-2xl font-bold text-amber-400">
                {secondaryValue}
              </div>
              <p className="text-xs text-gray-400 text-center mt-0.5">
                {secondaryLabel}
              </p>
            </div>
          </div>
        ) : (
          <>
            <div className="text-2xl font-bold text-amber-400">
              {value}
            </div>
            {description && (
              <p className="text-xs text-gray-400 text-center mt-0.5">
                {description}
              </p>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
}