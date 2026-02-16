import { AppSidebar } from "@/lib/features/dashboard/sidebar";
import { DashboardBreadcrumb } from "./dashboard-breadcrumb";
import { Separator } from "@/components/ui/separator";
import { DashboardCommonCopyProvider } from "@/lib/features/dashboard/i18n/common/common-copy-provider";
import { getDashboardCommonDictionary } from "@/lib/features/dashboard/i18n/common/get-common-dictionary";
import { resolveRequestDashboardLocale } from "@/lib/features/dashboard/i18n/common/resolve-request-locale";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";

export default async function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const locale = await resolveRequestDashboardLocale();
  const commonCopy = await getDashboardCommonDictionary(locale);

  return (
    <TooltipProvider>
      <DashboardCommonCopyProvider copy={commonCopy}>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset className="h-screen overflow-y-auto">
            <header className="bg-background/95 sticky top-0 z-20 flex h-16 shrink-0 items-center gap-2 border-b backdrop-blur supports-backdrop-filter:bg-background/80 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
              <div className="flex items-center gap-2 px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                  orientation="vertical"
                  className="mr-2 data-[orientation=vertical]:h-4"
                />
                <DashboardBreadcrumb />
              </div>
            </header>
            <main className="flex flex-1 flex-col gap-4 p-4">{children}</main>
          </SidebarInset>
        </SidebarProvider>
      </DashboardCommonCopyProvider>
    </TooltipProvider>
  );
}
