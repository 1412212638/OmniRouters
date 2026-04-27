/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useMemo, useRef, useState } from 'react';
import {
  Banner,
  Button,
  Card,
  Col,
  Form,
  Row,
  Spin,
  TagInput,
  TabPane,
  Tabs,
  Typography,
} from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

import { API, showError, showSuccess, toBoolean } from '../../helpers';

const { Text } = Typography;

const TEMPLATE_OPTION_KEYS = [
  'EmailVerificationSubjectTemplate',
  'EmailVerificationContentTemplate',
  'PasswordResetSubjectTemplate',
  'PasswordResetContentTemplate',
  'QuotaWarningSubjectTemplate',
  'QuotaWarningContentTemplate',
  'SubscriptionQuotaWarningSubjectTemplate',
  'SubscriptionQuotaWarningContentTemplate',
  'TopUpSuccessSubjectTemplate',
  'TopUpSuccessContentTemplate',
];

const TEMPLATE_PRESETS_ZH = {
  EmailVerificationSubjectTemplate: '{{system_name}}邮箱验证邮件',
  EmailVerificationContentTemplate:
    '<p>您好，您正在进行 {{system_name}} 邮箱验证。</p><p>您的验证码为：<strong>{{code}}</strong></p><p>验证码在 {{valid_time}} 内有效，如果不是本人操作，请忽略此邮件。</p>',
  PasswordResetSubjectTemplate: '{{system_name}}密码重置',
  PasswordResetContentTemplate:
    "<p>您好，您正在进行 {{system_name}} 密码重置。</p><p>点击 <a href='{{link}}'>此处</a> 进行密码重置。</p><p>如果链接无法点击，请尝试将下面的链接复制到浏览器中打开：<br>{{link}}</p><p>重置链接在 {{valid_time}} 内有效，如果不是本人操作，请忽略此邮件。</p>",
  QuotaWarningSubjectTemplate: '您的额度即将用尽',
  QuotaWarningContentTemplate:
    "<p>您好，</p><p>您的额度即将用尽。</p><p>当前剩余额度为 <strong>{{remaining_quota}}</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
  SubscriptionQuotaWarningSubjectTemplate: '您的订阅额度即将用尽',
  SubscriptionQuotaWarningContentTemplate:
    "<p>您好，</p><p>您的订阅额度即将用尽。</p><p>当前剩余额度为 <strong>{{remaining_quota}}</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
  TopUpSuccessSubjectTemplate: '{{system_name}}充值成功通知',
  TopUpSuccessContentTemplate:
    '<p>您好{{username}}，</p><p>您的充值已经到账。</p><p>支付金额：<strong>{{amount}}</strong></p><p>到账额度：<strong>{{quota}}</strong></p><p>当前余额：<strong>{{balance}}</strong></p><p>支付方式：{{payment_method}}</p><p>订单号：{{trade_no}}</p><p>到账时间：{{paid_at}}</p>',
};

const TEMPLATE_PRESETS_EN = {
  EmailVerificationSubjectTemplate: '{{system_name}} Email Verification',
  EmailVerificationContentTemplate:
    '<p>Hello,</p><p>You are verifying the email address for {{system_name}}.</p><p>Your verification code is: <strong>{{code}}</strong></p><p>This code is valid for {{valid_time}}. If this was not you, please ignore this email.</p>',
  PasswordResetSubjectTemplate: '{{system_name}} Password Reset',
  PasswordResetContentTemplate:
    "<p>Hello,</p><p>You requested a password reset for {{system_name}}.</p><p>Click <a href='{{link}}'>here</a> to reset your password.</p><p>If the link does not open, copy and paste this URL into your browser:<br>{{link}}</p><p>This reset link is valid for {{valid_time}}. If this was not you, please ignore this email.</p>",
  QuotaWarningSubjectTemplate: 'Your quota is running low',
  QuotaWarningContentTemplate:
    "<p>Hello,</p><p>Your quota is running low</p><p>Your remaining quota is <strong>{{remaining_quota}}</strong>.</p><p>To avoid interruption, please top up in time:</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
  SubscriptionQuotaWarningSubjectTemplate:
    'Your subscription quota is running low',
  SubscriptionQuotaWarningContentTemplate:
    "<p>Hello,</p><p>Your subscription quota is running low</p><p>Your remaining quota is <strong>{{remaining_quota}}</strong>.</p><p>To avoid interruption, please top up in time:</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
  TopUpSuccessSubjectTemplate: '{{system_name}} Top-up Successful',
  TopUpSuccessContentTemplate:
    '<p>Hello {{username}},</p><p>Your top-up has been completed.</p><p>Amount: <strong>{{amount}}</strong></p><p>Quota added: <strong>{{quota}}</strong></p><p>Current balance: <strong>{{balance}}</strong></p><p>Payment method: {{payment_method}}</p><p>Trade no: {{trade_no}}</p><p>Paid at: {{paid_at}}</p>',
};

const getTemplatePresets = (language) =>
  language === 'en' ? TEMPLATE_PRESETS_EN : TEMPLATE_PRESETS_ZH;

const DEFAULT_INPUTS = {
  SMTPServer: '',
  SMTPPort: '',
  SMTPAccount: '',
  SMTPFrom: '',
  SMTPToken: '',
  SMTPSSLEnabled: false,
  SMTPForceAuthLogin: false,
  EmailLanguage: 'zh',
  EmailDomainRestrictionEnabled: false,
  EmailAliasRestrictionEnabled: false,
  TopUpSuccessEmailEnabled: false,
  ...TEMPLATE_PRESETS_ZH,
};

const booleanKeys = new Set([
  'SMTPSSLEnabled',
  'SMTPForceAuthLogin',
  'EmailDomainRestrictionEnabled',
  'EmailAliasRestrictionEnabled',
  'TopUpSuccessEmailEnabled',
]);

const templateGroups = [
  {
    title: '注册 / 邮箱绑定验证码',
    description: '用于注册、绑定邮箱等验证码邮件',
    subjectKey: 'EmailVerificationSubjectTemplate',
    contentKey: 'EmailVerificationContentTemplate',
    variables: '{{system_name}}, {{code}}, {{valid_minutes}}, {{valid_time}}',
  },
  {
    title: '找回密码 / 重置密码',
    description: '用于用户通过邮箱找回密码',
    subjectKey: 'PasswordResetSubjectTemplate',
    contentKey: 'PasswordResetContentTemplate',
    variables: '{{system_name}}, {{link}}, {{valid_minutes}}, {{valid_time}}',
  },
  {
    title: '普通额度提醒',
    description: '用于用户普通额度低于提醒阈值时发送',
    subjectKey: 'QuotaWarningSubjectTemplate',
    contentKey: 'QuotaWarningContentTemplate',
    variables: '{{system_name}}, {{remaining_quota}}, {{top_up_link}}',
  },
  {
    title: '订阅额度提醒',
    description: '用于订阅套餐额度低于提醒阈值时发送',
    subjectKey: 'SubscriptionQuotaWarningSubjectTemplate',
    contentKey: 'SubscriptionQuotaWarningContentTemplate',
    variables: '{{system_name}}, {{remaining_quota}}, {{top_up_link}}',
  },
  {
    title: '充值成功通知',
    description: '用于易支付和 Stripe 普通充值成功到账后发送',
    subjectKey: 'TopUpSuccessSubjectTemplate',
    contentKey: 'TopUpSuccessContentTemplate',
    variables:
      '{{system_name}}, {{username}}, {{amount}}, {{quota}}, {{balance}}, {{trade_no}}, {{payment_method}}, {{payment_provider}}, {{paid_at}}',
  },
];

const getPreviewValues = (language) => ({
  system_name: 'OmniRouters',
  code: '123456',
  valid_minutes: '10',
  valid_time: language === 'en' ? '10 minutes' : '10 分钟',
  link: 'https://example.com/user/reset?email=user@example.com&token=sample',
  remaining_quota: language === 'en' ? '$1.23' : '1.23 美元',
  top_up_link: 'https://example.com/console/topup',
  username: language === 'en' ? 'Alex' : '张三',
  amount: language === 'en' ? '10.00' : '10.00',
  quota: language === 'en' ? '$10.00' : '10.00 美元',
  balance: language === 'en' ? '$18.88' : '18.88 美元',
  trade_no: 'topup_202604270001',
  payment_method: language === 'en' ? 'Stripe' : 'Stripe',
  payment_provider: language === 'en' ? 'stripe' : 'stripe',
  paid_at: language === 'en' ? '2026-04-27 20:30:00' : '2026-04-27 20:30:00',
});

const renderTemplate = (template, values) =>
  Object.entries(values).reduce(
    (result, [key, value]) =>
      result.replaceAll(`{{${key}}}`, String(value ?? '')),
    template || '',
  );

const MailSetting = () => {
  const { t } = useTranslation();
  const [inputs, setInputs] = useState(DEFAULT_INPUTS);
  const [originInputs, setOriginInputs] = useState(DEFAULT_INPUTS);
  const [loading, setLoading] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const [emailDomainWhitelist, setEmailDomainWhitelist] = useState([]);
  const [originEmailDomainWhitelist, setOriginEmailDomainWhitelist] = useState(
    [],
  );
  const [emailToAdd, setEmailToAdd] = useState('');
  const formApiRef = useRef(null);

  const currentTemplatePresets = useMemo(
    () => getTemplatePresets(inputs.EmailLanguage),
    [inputs.EmailLanguage],
  );

  const getEffectiveTemplate = (key) =>
    inputs[key] || currentTemplatePresets[key] || '';

  const getOptions = async () => {
    setLoading(true);
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (!success) {
      showError(message);
      setLoading(false);
      return;
    }

    const rawOptions = {};
    data.forEach((item) => {
      rawOptions[item.key] = item.value;
    });

    const language = rawOptions.EmailLanguage || DEFAULT_INPUTS.EmailLanguage;
    const presets = getTemplatePresets(language);
    const nextInputs = {
      ...DEFAULT_INPUTS,
      ...presets,
      EmailLanguage: language,
    };

    Object.keys(nextInputs).forEach((key) => {
      if (!(key in rawOptions)) return;
      if (TEMPLATE_OPTION_KEYS.includes(key)) {
        nextInputs[key] = rawOptions[key] || presets[key] || '';
      } else {
        nextInputs[key] = booleanKeys.has(key)
          ? toBoolean(rawOptions[key])
          : rawOptions[key] || '';
      }
    });

    const nextWhitelist = rawOptions.EmailDomainWhitelist
      ? rawOptions.EmailDomainWhitelist.split(',').filter(Boolean)
      : [];
    setEmailDomainWhitelist(nextWhitelist);
    setOriginEmailDomainWhitelist(nextWhitelist);
    setInputs(nextInputs);
    setOriginInputs(nextInputs);
    formApiRef.current?.setValues(nextInputs);
    setIsLoaded(true);
    setLoading(false);
  };

  useEffect(() => {
    getOptions();
  }, []);

  const updateOptions = async (options) => {
    if (options.length === 0) return true;

    setLoading(true);
    try {
      const results = await Promise.all(
        options.map((opt) =>
          API.put('/api/option/', {
            key: opt.key,
            value:
              typeof opt.value === 'boolean' ? opt.value.toString() : opt.value,
          }),
        ),
      );
      const errorResults = results.filter((res) => !res.data.success);
      errorResults.forEach((res) => showError(res.data.message));
      if (errorResults.length > 0) {
        setLoading(false);
        return false;
      }

      showSuccess(t('更新成功'));
      const nextInputs = { ...inputs };
      options.forEach((opt) => {
        if (opt.key in nextInputs) {
          nextInputs[opt.key] = opt.value;
        }
      });
      setInputs(nextInputs);
      setOriginInputs(nextInputs);
      setLoading(false);
      return true;
    } catch (error) {
      showError(t('更新失败'));
      setLoading(false);
      return false;
    }
  };

  const submitSMTP = async () => {
    const optionKeys = [
      'SMTPServer',
      'SMTPAccount',
      'SMTPFrom',
      'SMTPSSLEnabled',
      'SMTPForceAuthLogin',
      'EmailLanguage',
    ];
    const options = optionKeys
      .filter((key) => originInputs[key] !== inputs[key])
      .map((key) => ({ key, value: inputs[key] }));

    if (originInputs.SMTPPort !== inputs.SMTPPort && inputs.SMTPPort !== '') {
      options.push({ key: 'SMTPPort', value: inputs.SMTPPort });
    }
    if (
      originInputs.SMTPToken !== inputs.SMTPToken &&
      inputs.SMTPToken !== ''
    ) {
      options.push({ key: 'SMTPToken', value: inputs.SMTPToken });
    }

    await updateOptions(options);
  };

  const submitEmailDomainSettings = async () => {
    if (
      inputs.EmailDomainRestrictionEnabled &&
      emailDomainWhitelist.length === 0
    ) {
      showError(t('无法启用邮箱域名限制，请先填入限制的邮箱域名！'));
      return;
    }

    const options = [
      'EmailDomainRestrictionEnabled',
      'EmailAliasRestrictionEnabled',
    ]
      .filter((key) => originInputs[key] !== inputs[key])
      .map((key) => ({ key, value: inputs[key] }));

    const whitelistValue = emailDomainWhitelist.join(',');
    if (originEmailDomainWhitelist.join(',') !== whitelistValue) {
      options.push({ key: 'EmailDomainWhitelist', value: whitelistValue });
    }

    if (await updateOptions(options)) {
      setOriginEmailDomainWhitelist([...emailDomainWhitelist]);
    }
  };

  const submitTemplates = async () => {
    const options = ['TopUpSuccessEmailEnabled', ...TEMPLATE_OPTION_KEYS]
      .filter((key) => originInputs[key] !== inputs[key])
      .map((key) => ({
        key,
        value: booleanKeys.has(key) ? inputs[key] : inputs[key] || '',
      }));
    await updateOptions(options);
  };

  const restoreDefaultTemplate = (subjectKey, contentKey) => {
    const presets = getTemplatePresets(inputs.EmailLanguage);
    const nextInputs = {
      ...inputs,
      [subjectKey]: presets[subjectKey],
      [contentKey]: presets[contentKey],
    };
    setInputs(nextInputs);
    formApiRef.current?.setValues(nextInputs);
  };

  const handleAddEmail = () => {
    if (!emailToAdd || emailToAdd.trim() === '') return;
    const domain = emailToAdd.trim().toLowerCase();
    const domainRegex =
      /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/;
    if (!domainRegex.test(domain)) {
      showError(t('邮箱域名格式不正确，请输入有效的域名，如 gmail.com'));
      return;
    }
    if (emailDomainWhitelist.includes(domain)) {
      showError(t('该域名已存在于白名单中'));
      return;
    }
    setEmailDomainWhitelist([...emailDomainWhitelist, domain]);
    setEmailToAdd('');
    showSuccess(t('已添加到白名单'));
  };

  const handleFormChange = (values) => {
    setInputs({ ...inputs, ...values });
  };

  const handleCheckboxChange = (key, event) => {
    setInputs((prev) => ({ ...prev, [key]: event.target.checked }));
  };

  const renderTemplatePane = (group) => {
    const previewValues = getPreviewValues(inputs.EmailLanguage);
    const previewSubject = renderTemplate(
      getEffectiveTemplate(group.subjectKey),
      previewValues,
    );
    const previewContent = renderTemplate(
      getEffectiveTemplate(group.contentKey),
      previewValues,
    );

    return (
      <div style={{ paddingTop: 12 }}>
        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            gap: 12,
            alignItems: 'flex-start',
            marginBottom: 12,
          }}
        >
          <div>
            <Text type='tertiary'>{t(group.description)}</Text>
            <div>
              <Text type='tertiary'>
                {t('可用变量')}：{group.variables}
              </Text>
            </div>
          </div>
          <Button
            size='small'
            type='tertiary'
            onClick={() =>
              restoreDefaultTemplate(group.subjectKey, group.contentKey)
            }
          >
            {t('恢复默认模板')}
          </Button>
        </div>
        {group.subjectKey === 'TopUpSuccessSubjectTemplate' && (
          <Form.Checkbox
            field='TopUpSuccessEmailEnabled'
            noLabel
            style={{ marginBottom: 12 }}
            onChange={(event) =>
              handleCheckboxChange('TopUpSuccessEmailEnabled', event)
            }
          >
            {t('启用充值成功邮件通知（仅易支付和 Stripe 普通充值）')}
          </Form.Checkbox>
        )}
        <Form.Input field={group.subjectKey} label={t('邮件标题')} />
        <Form.TextArea
          field={group.contentKey}
          label={t('邮件内容')}
          autosize
        />
        <div
          style={{
            marginTop: 12,
            padding: 12,
            border: '1px dashed var(--semi-color-border)',
            borderRadius: 8,
            background: 'var(--semi-color-fill-0)',
          }}
        >
          <Text strong>{t('预览')}</Text>
          <div style={{ marginTop: 8 }}>
            <Text type='tertiary'>{t('标题')}：</Text>
            <Text>{previewSubject}</Text>
          </div>
          <div
            style={{
              marginTop: 8,
              padding: 12,
              borderRadius: 8,
              background: 'var(--semi-color-bg-0)',
            }}
            dangerouslySetInnerHTML={{ __html: previewContent }}
          />
        </div>
      </div>
    );
  };

  return (
    <div className='mail-setting'>
      {isLoaded ? (
        <Form
          initValues={inputs}
          onValueChange={handleFormChange}
          getFormApi={(api) => (formApiRef.current = api)}
        >
          <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
            <Card>
              <Form.Section text={t('配置 SMTP')}>
                <Text>{t('用以支持系统的邮件发送')}</Text>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                >
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='SMTPServer'
                      label={t('SMTP 服务器地址')}
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPPort' label={t('SMTP 端口')} />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPAccount' label={t('SMTP 账户')} />
                  </Col>
                </Row>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  style={{ marginTop: 16 }}
                >
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPFrom' label={t('SMTP 发送者邮箱')} />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='SMTPToken'
                      label={t('SMTP 访问凭证')}
                      type='password'
                      placeholder={t('敏感信息不会发送到前端显示')}
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Select
                      field='EmailLanguage'
                      label={t('邮件语言')}
                      optionList={[
                        { label: t('中文'), value: 'zh' },
                        { label: 'English', value: 'en' },
                      ]}
                    />
                  </Col>
                </Row>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  style={{ marginTop: 16 }}
                >
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Checkbox
                      field='SMTPSSLEnabled'
                      noLabel
                      onChange={(event) =>
                        handleCheckboxChange('SMTPSSLEnabled', event)
                      }
                    >
                      {t('启用SMTP SSL')}
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='SMTPForceAuthLogin'
                      noLabel
                      onChange={(event) =>
                        handleCheckboxChange('SMTPForceAuthLogin', event)
                      }
                    >
                      {t('强制使用 AUTH LOGIN')}
                    </Form.Checkbox>
                  </Col>
                </Row>
                <Button onClick={submitSMTP}>{t('保存 SMTP 设置')}</Button>
              </Form.Section>
            </Card>

            <Card>
              <Form.Section text={t('邮箱域名限制')}>
                <Text>{t('用以防止恶意用户利用临时邮箱批量注册')}</Text>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  style={{ marginTop: 12 }}
                >
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='EmailDomainRestrictionEnabled'
                      noLabel
                      onChange={(event) =>
                        handleCheckboxChange(
                          'EmailDomainRestrictionEnabled',
                          event,
                        )
                      }
                    >
                      {t('启用邮箱域名白名单')}
                    </Form.Checkbox>
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='EmailAliasRestrictionEnabled'
                      noLabel
                      onChange={(event) =>
                        handleCheckboxChange(
                          'EmailAliasRestrictionEnabled',
                          event,
                        )
                      }
                    >
                      {t('启用邮箱别名限制')}
                    </Form.Checkbox>
                  </Col>
                </Row>
                <TagInput
                  value={emailDomainWhitelist}
                  onChange={setEmailDomainWhitelist}
                  placeholder={t('输入域名后回车')}
                  style={{ width: '100%', marginTop: 16 }}
                />
                <Form.Input
                  placeholder={t('输入要添加的邮箱域名')}
                  value={emailToAdd}
                  onChange={(value) => setEmailToAdd(value)}
                  style={{ marginTop: 16 }}
                  suffix={
                    <Button
                      theme='solid'
                      type='primary'
                      onClick={handleAddEmail}
                    >
                      {t('添加')}
                    </Button>
                  }
                  onEnterPress={handleAddEmail}
                />
                <Button
                  onClick={submitEmailDomainSettings}
                  style={{ marginTop: 10 }}
                >
                  {t('保存邮箱域名限制')}
                </Button>
              </Form.Section>
            </Card>

            <Card>
              <Form.Section text={t('发件模板')}>
                <Banner
                  type='info'
                  description={t(
                    '这里展示的是当前生效模板；后台没有自定义模板时，会自动显示当前语言的内置默认模板。内容支持 HTML。',
                  )}
                  style={{ marginBottom: 16 }}
                />
                <Tabs type='line' collapsible>
                  {templateGroups.map((group) => (
                    <TabPane
                      itemKey={group.subjectKey}
                      key={group.subjectKey}
                      tab={t(group.title)}
                    >
                      {renderTemplatePane(group)}
                    </TabPane>
                  ))}
                </Tabs>
                <Button onClick={submitTemplates} style={{ marginTop: 16 }}>
                  {t('保存发件模板')}
                </Button>
              </Form.Section>
            </Card>
          </div>
        </Form>
      ) : (
        <div
          style={{
            display: 'flex',
            justifyContent: 'center',
            alignItems: 'center',
            height: '60vh',
          }}
        >
          <Spin size='large' spinning={loading} />
        </div>
      )}
    </div>
  );
};

export default MailSetting;
