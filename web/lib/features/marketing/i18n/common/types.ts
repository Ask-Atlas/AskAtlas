export interface MarketingLinkCopy {
  label: string;
  href: string;
}

export interface MarketingCommonDictionary {
  nav: {
    links: MarketingLinkCopy[];
    loginCta: string;
    primaryCta: string;
    dashboardCta: string;
  };
  footer: {
    links: MarketingLinkCopy[];
    socialLinks: MarketingLinkCopy[];
    privacyLabel: string;
    brandWordmark: string;
  };
}
