import { useMemo, useState } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Loader2, Search, Send, X } from "lucide-react";
import { useTranslation } from "react-i18next";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { TagInput } from "@/components/tag-input";
import { SettingsSection } from "../components/settings-section";
import { searchMarketingEmailUsers, sendMarketingEmail } from "../api";
import { useResetForm } from "../hooks/use-reset-form";
import { useUpdateOption } from "../hooks/use-update-option";
import type { MarketingEmailUser } from "../types";

const templateOptionKeys = [
  "EmailVerificationSubjectTemplate",
  "EmailVerificationContentTemplate",
  "PasswordResetSubjectTemplate",
  "PasswordResetContentTemplate",
  "QuotaWarningSubjectTemplate",
  "QuotaWarningContentTemplate",
  "SubscriptionQuotaWarningSubjectTemplate",
  "SubscriptionQuotaWarningContentTemplate",
  "TopUpSuccessSubjectTemplate",
  "TopUpSuccessContentTemplate",
  "MarketingEmailSubjectTemplate",
  "MarketingEmailContentTemplate",
] as const;

type TemplateOptionKey = (typeof templateOptionKeys)[number];

type TemplateGroup = {
  id: string;
  tab: string;
  title: string;
  description: string;
  subjectKey: TemplateOptionKey;
  contentKey: TemplateOptionKey;
  variables: string[];
  marketing?: boolean;
};

export const emailTemplatePresets = {
  zh: {
    EmailVerificationSubjectTemplate: "{{system_name}} Email Verification",
    EmailVerificationContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Thanks for joining {{system_name}}. Your verification code is:</p><p style="font-size: 42px; letter-spacing: 6px; font-weight: 700; color: #111827; margin: 28px 0;">{{code}}</p><p>Navigate back to your browser and enter the code. This code will expire in {{valid_time}}.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    PasswordResetSubjectTemplate: "{{system_name}} Password Reset",
    PasswordResetContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>We received a request to reset the password for your {{system_name}} account.</p><p><a href="{{link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Reset Password</a></p><p>If the button does not work, copy and paste this link into your browser:</p><p><a href="{{link}}">{{link}}</a></p><p>This reset link will expire in {{valid_time}}. If you did not request this, please ignore this email.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    QuotaWarningSubjectTemplate: "{{system_name}} Quota Reminder",
    QuotaWarningContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Your available quota is running low.</p><p>Current remaining quota: <strong>{{remaining_quota}}</strong></p><p>To avoid service interruption, please top up when convenient.</p><p><a href="{{top_up_link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Top Up Now</a></p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    SubscriptionQuotaWarningSubjectTemplate:
      "{{system_name}} Subscription Quota Reminder",
    SubscriptionQuotaWarningContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Your subscription quota is running low.</p><p>Current remaining subscription quota: <strong>{{remaining_quota}}</strong></p><p>To keep your workflow uninterrupted, please review your subscription or top up.</p><p><a href="{{top_up_link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Top Up Now</a></p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    TopUpSuccessSubjectTemplate: "{{system_name}} Top-up Successful",
    TopUpSuccessContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello {{username}},</p><p>Your top-up has been completed.</p><p>Amount: <strong>{{amount}}</strong></p><p>Quota added: <strong>{{quota}}</strong></p><p>Current balance: <strong>{{balance}}</strong></p><p>Payment method: {{payment_method}}</p><p>Trade no: {{trade_no}}</p><p>Paid at: {{paid_at}}</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    MarketingEmailSubjectTemplate: "{{system_name}} Update",
    MarketingEmailContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello {{display_name}},</p><p>Here is the latest update from {{system_name}}.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
  },
  en: {
    EmailVerificationSubjectTemplate: "{{system_name}} Email Verification",
    EmailVerificationContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Thanks for joining {{system_name}}. Your verification code is:</p><p style="font-size: 42px; letter-spacing: 6px; font-weight: 700; color: #111827; margin: 28px 0;">{{code}}</p><p>Navigate back to your browser and enter the code. This code will expire in {{valid_time}}.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    PasswordResetSubjectTemplate: "{{system_name}} Password Reset",
    PasswordResetContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>We received a request to reset the password for your {{system_name}} account.</p><p><a href="{{link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Reset Password</a></p><p>If the button does not work, copy and paste this link into your browser:</p><p><a href="{{link}}">{{link}}</a></p><p>This reset link will expire in {{valid_time}}. If you did not request this, please ignore this email.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    QuotaWarningSubjectTemplate: "{{system_name}} Quota Reminder",
    QuotaWarningContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Your available quota is running low.</p><p>Current remaining quota: <strong>{{remaining_quota}}</strong></p><p>To avoid service interruption, please top up when convenient.</p><p><a href="{{top_up_link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Top Up Now</a></p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    SubscriptionQuotaWarningSubjectTemplate:
      "{{system_name}} Subscription Quota Reminder",
    SubscriptionQuotaWarningContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello!</p><p>Your subscription quota is running low.</p><p>Current remaining subscription quota: <strong>{{remaining_quota}}</strong></p><p>To keep your workflow uninterrupted, please review your subscription or top up.</p><p><a href="{{top_up_link}}" style="display: inline-block; padding: 12px 18px; background: #111827; color: #ffffff; border-radius: 8px; text-decoration: none; font-weight: 700;">Top Up Now</a></p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    TopUpSuccessSubjectTemplate: "{{system_name}} Top-up Successful",
    TopUpSuccessContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello {{username}},</p><p>Your top-up has been completed.</p><p>Amount: <strong>{{amount}}</strong></p><p>Quota added: <strong>{{quota}}</strong></p><p>Current balance: <strong>{{balance}}</strong></p><p>Payment method: {{payment_method}}</p><p>Trade no: {{trade_no}}</p><p>Paid at: {{paid_at}}</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
    MarketingEmailSubjectTemplate: "{{system_name}} Update",
    MarketingEmailContentTemplate:
      '<div style="font-family: Arial, sans-serif; color: #1f2937; line-height: 1.6; max-width: 560px; margin: 0 auto; padding: 32px 24px;"><div style="display: flex; align-items: center; gap: 12px; margin-bottom: 32px;"><img src="https://cos.frostai.cn/omnirouters/logo/logo.png" alt="{{system_name}}" style="width: 44px; height: 44px;"><h1 style="margin: 0; font-size: 26px; color: #111827;">{{system_name}}</h1></div><p style="font-size: 18px;">Hello {{display_name}},</p><p>Here is the latest update from {{system_name}}.</p><p>If you have any questions or concerns, please contact us at <a href="mailto:support@omnirouters.com.com">support@omnirouters.com.com</a>.</p><p style="margin-top: 28px;">Thank you,<br>The {{system_name}} Team</p></div>',
  },
} as const;

export type EmailLanguage = keyof typeof emailTemplatePresets;

const createEmailSchema = (t: (key: string) => string) =>
  z.object({
    SMTPServer: z.string(),
    SMTPPort: z.string().refine((value) => {
      const trimmed = value.trim();
      if (!trimmed) return true;
      return /^\d+$/.test(trimmed);
    }, t("Port must be a positive integer")),
    SMTPAccount: z.string(),
    SMTPFrom: z.string().refine((value) => {
      const trimmed = value.trim();
      if (!trimmed) return true;
      return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(trimmed);
    }, t("Enter a valid email or leave blank")),
    SMTPToken: z.string(),
    SMTPSSLEnabled: z.boolean(),
    SMTPForceAuthLogin: z.boolean(),
    EmailLanguage: z.enum(["zh", "en"]),
    EmailDomainRestrictionEnabled: z.boolean(),
    EmailAliasRestrictionEnabled: z.boolean(),
    EmailDomainWhitelist: z.array(z.string()),
    TopUpSuccessEmailEnabled: z.boolean(),
    EmailVerificationSubjectTemplate: z.string(),
    EmailVerificationContentTemplate: z.string(),
    PasswordResetSubjectTemplate: z.string(),
    PasswordResetContentTemplate: z.string(),
    QuotaWarningSubjectTemplate: z.string(),
    QuotaWarningContentTemplate: z.string(),
    SubscriptionQuotaWarningSubjectTemplate: z.string(),
    SubscriptionQuotaWarningContentTemplate: z.string(),
    TopUpSuccessSubjectTemplate: z.string(),
    TopUpSuccessContentTemplate: z.string(),
    MarketingEmailSubjectTemplate: z.string(),
    MarketingEmailContentTemplate: z.string(),
  });

type EmailFormValues = z.infer<ReturnType<typeof createEmailSchema>>;

type EmailSettingsSectionProps = {
  defaultValues: EmailFormValues;
};

const templateGroups: TemplateGroup[] = [
  {
    id: "verification",
    tab: "Verification",
    title: "Registration / Email Verification",
    description: "Used for registration and email binding verification codes.",
    subjectKey: "EmailVerificationSubjectTemplate",
    contentKey: "EmailVerificationContentTemplate",
    variables: ["system_name", "code", "valid_minutes", "valid_time"],
  },
  {
    id: "password",
    tab: "Password",
    title: "Password Reset",
    description: "Used when users reset their password by email.",
    subjectKey: "PasswordResetSubjectTemplate",
    contentKey: "PasswordResetContentTemplate",
    variables: ["system_name", "link", "valid_minutes", "valid_time"],
  },
  {
    id: "quota",
    tab: "Quota",
    title: "Quota Reminder",
    description: "Used when a user quota falls below the reminder threshold.",
    subjectKey: "QuotaWarningSubjectTemplate",
    contentKey: "QuotaWarningContentTemplate",
    variables: ["system_name", "remaining_quota", "top_up_link"],
  },
  {
    id: "subscription",
    tab: "Subscription",
    title: "Subscription Reminder",
    description:
      "Used when subscription quota falls below the reminder threshold.",
    subjectKey: "SubscriptionQuotaWarningSubjectTemplate",
    contentKey: "SubscriptionQuotaWarningContentTemplate",
    variables: ["system_name", "remaining_quota", "top_up_link"],
  },
  {
    id: "topup",
    tab: "Top-up",
    title: "Top-up Success Notification",
    description:
      "Used after Epay and Stripe normal top-ups are successfully credited.",
    subjectKey: "TopUpSuccessSubjectTemplate",
    contentKey: "TopUpSuccessContentTemplate",
    variables: [
      "system_name",
      "username",
      "amount",
      "quota",
      "balance",
      "trade_no",
      "payment_method",
      "payment_provider",
      "paid_at",
    ],
  },
  {
    id: "marketing",
    tab: "Marketing",
    title: "Marketing Email",
    description:
      "Used for manual announcements or marketing emails to selected users.",
    subjectKey: "MarketingEmailSubjectTemplate",
    contentKey: "MarketingEmailContentTemplate",
    variables: ["system_name", "username", "display_name", "email", "user_id"],
    marketing: true,
  },
];

const booleanKeys = new Set<keyof EmailFormValues>([
  "SMTPSSLEnabled",
  "SMTPForceAuthLogin",
  "EmailDomainRestrictionEnabled",
  "EmailAliasRestrictionEnabled",
  "TopUpSuccessEmailEnabled",
]);

function normalizeDomainWhitelist(domains: string[]) {
  return domains.map((domain) => domain.trim().toLowerCase()).filter(Boolean);
}

function isValidDomain(domain: string) {
  return /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/.test(
    domain,
  );
}

function renderTemplate(template: string, values: Record<string, string>) {
  return Object.entries(values).reduce(
    (result, [key, value]) =>
      result.replaceAll(`{{${key}}}`, String(value ?? "")),
    template || "",
  );
}

function getPreviewValues(language: EmailLanguage) {
  return {
    system_name: "OmniRouters",
    code: "SLJ4J4",
    valid_minutes: "15",
    valid_time: language === "en" ? "15 minutes" : "15 minutes",
    link: "https://example.com/user/reset?email=user@example.com&token=sample",
    remaining_quota: "$1.23",
    top_up_link: "https://example.com/console/topup",
    username: "alex",
    display_name: "Alex Chen",
    email: "alex@example.com",
    user_id: "42",
    amount: "$10.00",
    quota: "$10.00",
    balance: "$18.88",
    trade_no: "topup_202604270001",
    payment_method: "Stripe",
    payment_provider: "stripe",
    paid_at: "2026-04-27 20:30:00",
  };
}

function getEffectiveTemplate(values: EmailFormValues, key: TemplateOptionKey) {
  const language = values.EmailLanguage;
  return values[key] || emailTemplatePresets[language][key] || "";
}

function formatUserOption(user: MarketingEmailUser) {
  const username = user.username || `#${user.id}`;
  const email = user.email || "no email";
  return `${username} · ${email}`;
}

export function EmailSettingsSection({
  defaultValues,
}: EmailSettingsSectionProps) {
  const { t } = useTranslation();
  const updateOption = useUpdateOption();
  const emailSchema = createEmailSchema(t);
  const [activeTemplate, setActiveTemplate] = useState(templateGroups[0].id);
  const [userKeyword, setUserKeyword] = useState("");
  const [userOptions, setUserOptions] = useState<MarketingEmailUser[]>([]);
  const [selectedUsers, setSelectedUsers] = useState<MarketingEmailUser[]>([]);
  const [searchingUsers, setSearchingUsers] = useState(false);
  const [sendingMarketing, setSendingMarketing] = useState(false);

  const form = useForm<EmailFormValues>({
    resolver: zodResolver(emailSchema),
    defaultValues,
  });

  useResetForm(form, defaultValues);

  const watchedValues = form.watch();

  const changedUpdates = useMemo(() => {
    const current = form.getValues();
    const updates: Array<{ key: string; value: string | boolean }> = [];

    const normalizedCurrent: EmailFormValues = {
      ...current,
      SMTPServer: current.SMTPServer.trim(),
      SMTPPort: current.SMTPPort.trim(),
      SMTPAccount: current.SMTPAccount.trim(),
      SMTPFrom: current.SMTPFrom.trim(),
      SMTPToken: current.SMTPToken.trim(),
      EmailDomainWhitelist: normalizeDomainWhitelist(
        current.EmailDomainWhitelist,
      ),
    };
    const normalizedInitial: EmailFormValues = {
      ...defaultValues,
      SMTPServer: defaultValues.SMTPServer.trim(),
      SMTPPort: defaultValues.SMTPPort.trim(),
      SMTPAccount: defaultValues.SMTPAccount.trim(),
      SMTPFrom: defaultValues.SMTPFrom.trim(),
      SMTPToken: defaultValues.SMTPToken.trim(),
      EmailDomainWhitelist: normalizeDomainWhitelist(
        defaultValues.EmailDomainWhitelist,
      ),
    };

    (Object.keys(normalizedCurrent) as Array<keyof EmailFormValues>).forEach(
      (key) => {
        if (key === "SMTPToken" && normalizedCurrent.SMTPToken === "") return;

        const currentValue = normalizedCurrent[key];
        const initialValue = normalizedInitial[key];
        const currentComparable = Array.isArray(currentValue)
          ? currentValue.join(",")
          : String(currentValue);
        const initialComparable = Array.isArray(initialValue)
          ? initialValue.join(",")
          : String(initialValue);

        if (currentComparable === initialComparable) return;

        if (key === "EmailDomainWhitelist") {
          updates.push({ key, value: currentComparable });
          return;
        }

        updates.push({
          key,
          value: booleanKeys.has(key)
            ? Boolean(currentValue)
            : String(currentValue ?? ""),
        });
      },
    );

    return updates;
  }, [defaultValues, form, watchedValues]);

  const onSubmit = async (values: EmailFormValues) => {
    const domains = normalizeDomainWhitelist(values.EmailDomainWhitelist);
    const invalidDomain = domains.find((domain) => !isValidDomain(domain));
    if (invalidDomain) {
      toast.error(
        t("Invalid email domain: {{domain}}", { domain: invalidDomain }),
      );
      return;
    }

    if (values.EmailDomainRestrictionEnabled && domains.length === 0) {
      toast.error(t("Add at least one domain before enabling the whitelist"));
      return;
    }

    if (changedUpdates.length === 0) {
      toast.info(t("No changes to save"));
      return;
    }

    for (const update of changedUpdates) {
      await updateOption.mutateAsync(update);
    }
  };

  const restoreTemplate = (group: TemplateGroup) => {
    const presets = emailTemplatePresets[watchedValues.EmailLanguage];
    form.setValue(group.subjectKey, presets[group.subjectKey], {
      shouldDirty: true,
    });
    form.setValue(group.contentKey, presets[group.contentKey], {
      shouldDirty: true,
    });
  };

  const handleSearchUsers = async () => {
    setSearchingUsers(true);
    try {
      const result = await searchMarketingEmailUsers({
        keyword: userKeyword.trim(),
        group: "",
        p: 1,
        page_size: 20,
      });
      if (!result.success) {
        toast.error(result.message || t("Failed to search users"));
        return;
      }
      setUserOptions(result.data?.items || []);
    } catch (error) {
      toast.error((error as Error)?.message || t("Failed to search users"));
    } finally {
      setSearchingUsers(false);
    }
  };

  const toggleSelectedUser = (user: MarketingEmailUser) => {
    setSelectedUsers((prev) => {
      if (prev.some((item) => item.id === user.id)) {
        return prev.filter((item) => item.id !== user.id);
      }
      return [...prev, user];
    });
  };

  const handleSendMarketingEmail = async () => {
    const values = form.getValues();
    const subjectTemplate = getEffectiveTemplate(
      values,
      "MarketingEmailSubjectTemplate",
    );
    const contentTemplate = getEffectiveTemplate(
      values,
      "MarketingEmailContentTemplate",
    );

    if (selectedUsers.length === 0) {
      toast.error(t("Select at least one recipient"));
      return;
    }
    if (!subjectTemplate.trim() || !contentTemplate.trim()) {
      toast.error(t("Marketing email subject and content are required"));
      return;
    }

    setSendingMarketing(true);
    try {
      const result = await sendMarketingEmail({
        user_ids: selectedUsers.map((user) => user.id),
        subject_template: subjectTemplate,
        content_template: contentTemplate,
      });
      if (!result.success) {
        toast.error(result.message || t("Failed to send marketing email"));
        return;
      }
      toast.success(
        t("Sent: {{sent}}, skipped: {{skipped}}, failed: {{failed}}", {
          sent: result.data?.sent || 0,
          skipped: result.data?.skipped || 0,
          failed: result.data?.failed || 0,
        }),
      );
    } catch (error) {
      toast.error(
        (error as Error)?.message || t("Failed to send marketing email"),
      );
    } finally {
      setSendingMarketing(false);
    }
  };

  const renderTemplateFields = (group: TemplateGroup) => {
    const previewValues = getPreviewValues(watchedValues.EmailLanguage);
    const previewSubject = renderTemplate(
      getEffectiveTemplate(watchedValues, group.subjectKey),
      previewValues,
    );
    const previewContent = renderTemplate(
      getEffectiveTemplate(watchedValues, group.contentKey),
      previewValues,
    );

    return (
      <div className="space-y-5">
        <div className="flex flex-col gap-3 rounded-lg border bg-muted/25 p-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="space-y-2">
            <div>
              <h3 className="text-sm font-semibold">{t(group.title)}</h3>
              <p className="text-muted-foreground mt-1 text-sm">
                {t(group.description)}
              </p>
            </div>
            <div className="flex flex-wrap gap-1.5">
              {group.variables.map((variable) => (
                <Badge key={variable} variant="secondary" className="font-mono">
                  {`{{${variable}}}`}
                </Badge>
              ))}
            </div>
          </div>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => restoreTemplate(group)}
          >
            {t("Restore default")}
          </Button>
        </div>

        {group.id === "topup" && (
          <FormField
            control={form.control}
            name="TopUpSuccessEmailEnabled"
            render={({ field }) => (
              <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                <div className="space-y-0.5">
                  <FormLabel>{t("Enable top-up success email")}</FormLabel>
                  <FormDescription>
                    {t("Only sent for Epay and Stripe normal top-ups")}
                  </FormDescription>
                </div>
                <FormControl>
                  <Switch
                    checked={field.value}
                    onCheckedChange={field.onChange}
                  />
                </FormControl>
              </FormItem>
            )}
          />
        )}

        <FormField
          control={form.control}
          name={group.subjectKey}
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("Email Subject")}</FormLabel>
              <FormControl>
                <Input {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <FormField
          control={form.control}
          name={group.contentKey}
          render={({ field }) => (
            <FormItem>
              <FormLabel>{t("Email Content")}</FormLabel>
              <FormControl>
                <Textarea rows={10} className="font-mono text-xs" {...field} />
              </FormControl>
              <FormDescription>
                {t(
                  "HTML content is supported. Variables are replaced when sent.",
                )}
              </FormDescription>
              <FormMessage />
            </FormItem>
          )}
        />

        <div className="rounded-lg border bg-background p-4">
          <div className="mb-3 flex items-center justify-between gap-3">
            <h4 className="text-sm font-semibold">{t("Preview")}</h4>
            <Badge variant="outline" className="font-mono">
              {previewSubject}
            </Badge>
          </div>
          <div
            className="prose prose-sm dark:prose-invert max-w-none overflow-hidden rounded-md border bg-white p-4 text-slate-900"
            dangerouslySetInnerHTML={{ __html: previewContent }}
          />
        </div>

        {group.marketing && (
          <div className="space-y-4 rounded-lg border bg-muted/25 p-4">
            <div>
              <h4 className="text-sm font-semibold">{t("Recipients")}</h4>
              <p className="text-muted-foreground mt-1 text-sm">
                {t("Search users, then select one or more recipients.")}
              </p>
            </div>

            <div className="flex flex-col gap-2 sm:flex-row">
              <Input
                value={userKeyword}
                onChange={(event) => setUserKeyword(event.target.value)}
                onKeyDown={(event) => {
                  if (event.key === "Enter") {
                    event.preventDefault();
                    handleSearchUsers();
                  }
                }}
                placeholder={t("Search username, email, or user ID")}
              />
              <Button
                type="button"
                variant="outline"
                onClick={handleSearchUsers}
                disabled={searchingUsers}
              >
                {searchingUsers ? (
                  <Loader2 className="size-4 animate-spin" />
                ) : (
                  <Search className="size-4" />
                )}
                {t("Search")}
              </Button>
            </div>

            {selectedUsers.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {selectedUsers.map((user) => (
                  <Badge key={user.id} variant="secondary" className="gap-1">
                    {formatUserOption(user)}
                    <button
                      type="button"
                      className="rounded-sm hover:bg-muted-foreground/10"
                      onClick={() => toggleSelectedUser(user)}
                    >
                      <X className="size-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            )}

            <div className="max-h-56 overflow-y-auto rounded-md border bg-background">
              {userOptions.length === 0 ? (
                <div className="text-muted-foreground p-4 text-sm">
                  {t("Search users to add recipients")}
                </div>
              ) : (
                userOptions.map((user) => {
                  const checked = selectedUsers.some(
                    (item) => item.id === user.id,
                  );
                  return (
                    <label
                      key={user.id}
                      className={cn(
                        "flex cursor-pointer items-center gap-3 border-b p-3 text-sm last:border-b-0 hover:bg-muted/50",
                        checked && "bg-muted/60",
                      )}
                    >
                      <Checkbox
                        checked={checked}
                        onCheckedChange={() => toggleSelectedUser(user)}
                      />
                      <span className="min-w-0 flex-1 truncate">
                        {formatUserOption(user)}
                      </span>
                      <span className="text-muted-foreground font-mono text-xs">
                        #{user.id}
                      </span>
                    </label>
                  );
                })
              )}
            </div>

            <Button
              type="button"
              onClick={handleSendMarketingEmail}
              disabled={sendingMarketing || selectedUsers.length === 0}
            >
              {sendingMarketing ? (
                <Loader2 className="size-4 animate-spin" />
              ) : (
                <Send className="size-4" />
              )}
              {t("Send Marketing Email")}
            </Button>
          </div>
        )}
      </div>
    );
  };

  return (
    <SettingsSection
      title={t("Email Settings")}
      description={t("Configure SMTP, sender templates, and marketing emails")}
    >
      <Form {...form}>
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          className="space-y-8"
          autoComplete="off"
        >
          <div className="space-y-4">
            <h3 className="text-sm font-semibold">{t("SMTP")}</h3>

            <FormField
              control={form.control}
              name="SMTPServer"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("SMTP Host")}</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="off"
                      placeholder={t("smtp.example.com")}
                      {...field}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <div className="grid gap-6 md:grid-cols-2">
              <FormField
                control={form.control}
                name="SMTPPort"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Port")}</FormLabel>
                    <FormControl>
                      <Input
                        autoComplete="off"
                        type="number"
                        placeholder="587"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      {t("Common ports include 25, 465, and 587")}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="EmailLanguage"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Email Language")}</FormLabel>
                    <Select value={field.value} onValueChange={field.onChange}>
                      <FormControl>
                        <SelectTrigger className="w-full">
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="zh">{t("Chinese")}</SelectItem>
                        <SelectItem value="en">English</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      {t(
                        "Used by built-in template presets and backend defaults",
                      )}
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <div className="grid gap-6 md:grid-cols-2">
              <FormField
                control={form.control}
                name="SMTPSSLEnabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>{t("Enable SSL/TLS")}</FormLabel>
                      <FormDescription>
                        {t("Use secure connection when sending emails")}
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="SMTPForceAuthLogin"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>{t("Force AUTH LOGIN")}</FormLabel>
                      <FormDescription>
                        {t("Force SMTP authentication using AUTH LOGIN method")}
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <div className="grid gap-6 md:grid-cols-2">
              <FormField
                control={form.control}
                name="SMTPAccount"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("Username")}</FormLabel>
                    <FormControl>
                      <Input
                        autoComplete="off"
                        placeholder={t("noreply@example.com")}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="SMTPFrom"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t("From Address")}</FormLabel>
                    <FormControl>
                      <Input
                        autoComplete="off"
                        placeholder={t("OmniRouters <noreply@example.com>")}
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="SMTPToken"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("Password / Access Token")}</FormLabel>
                  <FormControl>
                    <Input
                      autoComplete="off"
                      type="password"
                      placeholder={t("Enter new token to update")}
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    {t("Leave blank to keep the existing credential")}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className="space-y-4">
            <h3 className="text-sm font-semibold">
              {t("Email Domain Whitelist")}
            </h3>
            <div className="grid gap-6 md:grid-cols-2">
              <FormField
                control={form.control}
                name="EmailDomainRestrictionEnabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>{t("Enable whitelist")}</FormLabel>
                      <FormDescription>
                        {t("Only allow registration from listed domains")}
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="EmailAliasRestrictionEnabled"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between rounded-lg border p-4">
                    <div className="space-y-0.5">
                      <FormLabel>{t("Block email aliases")}</FormLabel>
                      <FormDescription>
                        {t("Restrict plus-address aliases when registering")}
                      </FormDescription>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="EmailDomainWhitelist"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t("Allowed Domains")}</FormLabel>
                  <FormControl>
                    <TagInput
                      value={field.value}
                      onChange={field.onChange}
                      placeholder={t("Add domains such as gmail.com")}
                    />
                  </FormControl>
                  <FormDescription>
                    {t("Press Enter or comma after each domain")}
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-semibold">{t("Sender Templates")}</h3>
              <p className="text-muted-foreground mt-1 text-sm">
                {t(
                  "These templates are the effective sender templates. Empty custom values fall back to the selected language preset.",
                )}
              </p>
            </div>

            <Tabs value={activeTemplate} onValueChange={setActiveTemplate}>
              <TabsList className="grid h-auto w-full grid-cols-2 md:grid-cols-3 xl:grid-cols-6">
                {templateGroups.map((group) => (
                  <TabsTrigger key={group.id} value={group.id}>
                    {t(group.tab)}
                  </TabsTrigger>
                ))}
              </TabsList>
              {templateGroups.map((group) => (
                <TabsContent key={group.id} value={group.id} className="mt-4">
                  {renderTemplateFields(group)}
                </TabsContent>
              ))}
            </Tabs>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Button type="submit" disabled={updateOption.isPending}>
              {updateOption.isPending ? (
                <Loader2 className="size-4 animate-spin" />
              ) : null}
              {updateOption.isPending
                ? t("Saving...")
                : t("Save Email Settings")}
            </Button>
            {changedUpdates.length > 0 && (
              <span className="text-muted-foreground text-sm">
                {t("{{count}} pending changes", {
                  count: changedUpdates.length,
                })}
              </span>
            )}
          </div>
        </form>
      </Form>
    </SettingsSection>
  );
}
