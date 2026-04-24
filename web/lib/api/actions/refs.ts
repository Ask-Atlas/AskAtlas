"use server";

import { serverApi } from "../server-client";
import { unwrap } from "../errors";
import type { RefsResolveRequest, RefsResolveResponse } from "../types";

export async function resolveRefs(
  body: RefsResolveRequest,
): Promise<RefsResolveResponse> {
  return unwrap(
    await serverApi.POST("/refs/resolve", { body }),
    "POST /refs/resolve",
  );
}
