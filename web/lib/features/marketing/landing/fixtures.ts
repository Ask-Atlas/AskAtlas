export interface FeaturedStudent {
  id: number;
  name: string;
  avatar: string;
  university: string;
}

export const FEATURED_STUDENTS: FeaturedStudent[] = [
  {
    id: 1,
    name: "Sarah Chen",
    avatar: "/avatars/sarah.png",
    university: "MIT",
  },
  {
    id: 2,
    name: "Marcus Johnson",
    avatar: "/avatars/marcus.png",
    university: "Stanford",
  },
  {
    id: 3,
    name: "Elena Rodriguez",
    avatar: "/avatars/elena.png",
    university: "Harvard",
  },
  {
    id: 4,
    name: "David Kim",
    avatar: "/avatars/david.png",
    university: "Berkeley",
  },
];

export const UNIVERSITIES: string[] = [
  "Harvard University",
  "Stanford University",
  "MIT",
  "UC Berkeley",
  "Oxford University",
  "Cambridge University",
  "Yale University",
  "Princeton University",
  "Columbia University",
  "University of Toronto",
  "NYU",
  "UCLA",
  "University of Michigan",
  "Cornell University",
  "Duke University",
  "Carnegie Mellon",
  "Georgia Tech",
  "Northwestern University",
];
