import { MarketingNavbar } from "@/lib/features/marketing/MarketingNavbar"

export default function MarketingLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div className="relative flex min-h-screen flex-col">
      <header className="sticky top-0 z-50 w-full bg-background/95 backdrop-blur">
        <MarketingNavbar />
      </header>
      
      <main className="flex-1">
        {children}
      </main>
      
      <footer className="border-t">
        <div className="container py-8">
          {/* Footer */}
        </div>
      </footer>
    </div>
  )
}