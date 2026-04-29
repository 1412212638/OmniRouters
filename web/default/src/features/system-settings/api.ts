import { api } from "@/lib/api";
import type {
  DeleteLogsResponse,
  FetchUpstreamRatiosRequest,
  SearchMarketingEmailUsersParams,
  SearchMarketingEmailUsersResponse,
  SendMarketingEmailRequest,
  SendMarketingEmailResponse,
  SystemOptionsResponse,
  UpdateOptionRequest,
  UpdateOptionResponse,
  UpstreamChannelsResponse,
  UpstreamRatiosResponse,
} from "./types";

export async function getSystemOptions() {
  const res = await api.get<SystemOptionsResponse>("/api/option/");
  return res.data;
}

export async function updateSystemOption(request: UpdateOptionRequest) {
  const res = await api.put<UpdateOptionResponse>("/api/option/", request);
  return res.data;
}

export async function deleteLogsBefore(targetTimestamp: number) {
  const res = await api.delete<DeleteLogsResponse>("/api/log/", {
    params: { target_timestamp: targetTimestamp },
  });
  return res.data;
}

export async function searchMarketingEmailUsers(
  params: SearchMarketingEmailUsersParams,
) {
  const res = await api.get<SearchMarketingEmailUsersResponse>(
    "/api/user/search",
    { params },
  );
  return res.data;
}

export async function sendMarketingEmail(request: SendMarketingEmailRequest) {
  const res = await api.post<SendMarketingEmailResponse>(
    "/api/option/marketing_email/send",
    request,
  );
  return res.data;
}

export async function resetModelRatios() {
  const res = await api.post<UpdateOptionResponse>(
    "/api/option/rest_model_ratio",
  );
  return res.data;
}

export async function getUpstreamChannels() {
  const res = await api.get<UpstreamChannelsResponse>(
    "/api/ratio_sync/channels",
  );
  return res.data;
}

export async function fetchUpstreamRatios(request: FetchUpstreamRatiosRequest) {
  const res = await api.post<UpstreamRatiosResponse>(
    "/api/ratio_sync/fetch",
    request,
  );
  return res.data;
}
