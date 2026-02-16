"use client";

import { useClerk, useUser } from "@clerk/nextjs";

interface NavUserProps {
  name: string;
  email: string;
  avatar: string;
}

export function useNavUser(initialUser: NavUserProps) {
  const { user, isLoaded } = useUser();
  const { signOut, openUserProfile } = useClerk();

  const displayUser =
    isLoaded && user
      ? {
          name: user.fullName || user.username || "User",
          email: user.primaryEmailAddress?.emailAddress || "",
          avatar: user.imageUrl,
        }
      : initialUser;

  return {
    user: displayUser,
    isLoading: !isLoaded,
    signOut,
    openUserProfile,
  };
}
