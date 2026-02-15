import { MarketingNavbar } from "@/lib/features/marketing/marketing-navbar";
import { Footer } from "@/lib/features/marketing/footer";
import { LandingCopyProvider } from "@/lib/features/marketing/landing/i18n/landing-copy-provider";
import { getLandingDictionary } from "@/lib/features/marketing/landing/i18n/get-landing-dictionary";
import { resolveRequestLandingLocale } from "@/lib/features/marketing/landing/i18n/resolve-request-locale";
import { CommonCopyProvider } from "@/lib/features/marketing/i18n/common/common-copy-provider";
import { getMarketingCommonDictionary } from "@/lib/features/marketing/i18n/common/get-common-dictionary";

export default async function MarketingLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const locale = await resolveRequestLandingLocale();
  const [landingCopy, commonCopy] = await Promise.all([
    getLandingDictionary(locale),
    getMarketingCommonDictionary(locale),
  ]);

  return (
    <CommonCopyProvider copy={commonCopy}>
      <LandingCopyProvider copy={landingCopy}>
        <div className="relative flex min-h-screen flex-col">
          <header className="sticky top-0 z-50 w-full">
            <MarketingNavbar />
          </header>

          <main className="flex-1">{children}</main>

          <div className="border-t">
            <Footer />
          </div>
        </div>
      </LandingCopyProvider>
    </CommonCopyProvider>
  );
}
