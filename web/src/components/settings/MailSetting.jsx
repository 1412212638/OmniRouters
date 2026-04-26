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

import React, { useEffect, useRef, useState } from 'react';
import { Banner, Button, Card, Col, Form, Row, Spin, Typography } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

import { API, showError, showSuccess, toBoolean } from '../../helpers';

const { Text } = Typography;

const TEMPLATE_PRESETS = {
  EmailVerificationSubjectTemplate: '{{system_name}}邮箱验证邮件',
  EmailVerificationContentTemplate:
    '<p>您好，您正在进行 {{system_name}} 邮箱验证。</p><p>您的验证码为：<strong>{{code}}</strong></p><p>验证码在 {{valid_time}} 内有效，如果不是本人操作，请忽略此邮件。</p>',
  PasswordResetSubjectTemplate: '{{system_name}}密码重置',
  PasswordResetContentTemplate:
    "<p>您好，您正在进行 {{system_name}} 密码重置。</p><p>点击 <a href='{{link}}'>此处</a> 进行密码重置。</p><p>如果链接无法点击，请复制下面的链接到浏览器打开：<br>{{link}}</p><p>重置链接在 {{valid_time}} 内有效，如果不是本人操作，请忽略此邮件。</p>",
  QuotaWarningSubjectTemplate: '您的额度即将用尽',
  QuotaWarningContentTemplate:
    "<p>您好，</p><p>您的额度即将用尽。</p><p>当前剩余额度为 <strong>{{remaining_quota}}</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
  SubscriptionQuotaWarningSubjectTemplate: '您的订阅额度即将用尽',
  SubscriptionQuotaWarningContentTemplate:
    "<p>您好，</p><p>您的订阅额度即将用尽。</p><p>当前剩余额度为 <strong>{{remaining_quota}}</strong>。</p><p>为了不影响您的使用，请及时充值：</p><p><a href='{{top_up_link}}'>{{top_up_link}}</a></p>",
};

const DEFAULT_INPUTS = {
  SMTPServer: '',
  SMTPPort: '',
  SMTPAccount: '',
  SMTPFrom: '',
  SMTPToken: '',
  SMTPSSLEnabled: false,
  SMTPForceAuthLogin: false,
  EmailLanguage: 'zh',
  ...Object.fromEntries(Object.keys(TEMPLATE_PRESETS).map((key) => [key, ''])),
};

const booleanKeys = new Set(['SMTPSSLEnabled', 'SMTPForceAuthLogin']);

const templateGroups = [
  {
    title: '邮箱验证码模板',
    subjectKey: 'EmailVerificationSubjectTemplate',
    contentKey: 'EmailVerificationContentTemplate',
    variables: '{{system_name}}, {{code}}, {{valid_minutes}}, {{valid_time}}',
  },
  {
    title: '密码重置模板',
    subjectKey: 'PasswordResetSubjectTemplate',
    contentKey: 'PasswordResetContentTemplate',
    variables: '{{system_name}}, {{link}}, {{valid_minutes}}, {{valid_time}}',
  },
  {
    title: '额度提醒模板',
    subjectKey: 'QuotaWarningSubjectTemplate',
    contentKey: 'QuotaWarningContentTemplate',
    variables: '{{system_name}}, {{remaining_quota}}, {{top_up_link}}',
  },
  {
    title: '订阅额度提醒模板',
    subjectKey: 'SubscriptionQuotaWarningSubjectTemplate',
    contentKey: 'SubscriptionQuotaWarningContentTemplate',
    variables: '{{system_name}}, {{remaining_quota}}, {{top_up_link}}',
  },
];

const MailSetting = () => {
  const { t } = useTranslation();
  const [inputs, setInputs] = useState(DEFAULT_INPUTS);
  const [originInputs, setOriginInputs] = useState(DEFAULT_INPUTS);
  const [loading, setLoading] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const formApiRef = useRef(null);

  const getOptions = async () => {
    setLoading(true);
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (!success) {
      showError(message);
      setLoading(false);
      return;
    }

    const nextInputs = { ...DEFAULT_INPUTS };
    data.forEach((item) => {
      if (!(item.key in nextInputs)) return;
      nextInputs[item.key] = booleanKeys.has(item.key)
        ? toBoolean(item.value)
        : item.value || '';
    });
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
    if (options.length === 0) return;

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
      if (errorResults.length === 0) {
        showSuccess(t('更新成功'));
        const nextInputs = { ...inputs };
        options.forEach((opt) => {
          nextInputs[opt.key] = opt.value;
        });
        setInputs(nextInputs);
        setOriginInputs(nextInputs);
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
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
    if (originInputs.SMTPToken !== inputs.SMTPToken && inputs.SMTPToken !== '') {
      options.push({ key: 'SMTPToken', value: inputs.SMTPToken });
    }

    await updateOptions(options);
  };

  const submitTemplates = async () => {
    const options = Object.keys(TEMPLATE_PRESETS)
      .filter((key) => originInputs[key] !== inputs[key])
      .map((key) => ({ key, value: inputs[key] || '' }));
    await updateOptions(options);
  };

  const fillPreset = (subjectKey, contentKey) => {
    const nextInputs = {
      ...inputs,
      [subjectKey]: TEMPLATE_PRESETS[subjectKey],
      [contentKey]: TEMPLATE_PRESETS[contentKey],
    };
    setInputs(nextInputs);
    formApiRef.current?.setValues(nextInputs);
  };

  const clearTemplate = (subjectKey, contentKey) => {
    const nextInputs = {
      ...inputs,
      [subjectKey]: '',
      [contentKey]: '',
    };
    setInputs(nextInputs);
    formApiRef.current?.setValues(nextInputs);
  };

  const handleFormChange = (values) => {
    setInputs({ ...inputs, ...values });
  };

  const handleCheckboxChange = (key, event) => {
    setInputs((prev) => ({ ...prev, [key]: event.target.checked }));
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
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPServer' label={t('SMTP 服务器地址')} />
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
              <Form.Section text={t('发件模板')}>
                <Banner
                  type='info'
                  description={t(
                    '模板为空时使用系统默认文案。内容支持 HTML，并可使用下方列出的变量。',
                  )}
                  style={{ marginBottom: 16 }}
                />
                <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                  {templateGroups.map((group) => (
                    <div
                      key={group.subjectKey}
                      style={{
                        border: '1px solid var(--semi-color-border)',
                        borderRadius: 8,
                        padding: 16,
                      }}
                    >
                      <div
                        style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          gap: 12,
                          alignItems: 'center',
                          marginBottom: 12,
                        }}
                      >
                        <div>
                          <Text strong>{t(group.title)}</Text>
                          <div>
                            <Text type='tertiary'>
                              {t('可用变量')}：{group.variables}
                            </Text>
                          </div>
                        </div>
                        <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                          <Button
                            size='small'
                            type='tertiary'
                            onClick={() =>
                              fillPreset(group.subjectKey, group.contentKey)
                            }
                          >
                            {t('填入默认模板')}
                          </Button>
                          <Button
                            size='small'
                            type='tertiary'
                            onClick={() =>
                              clearTemplate(group.subjectKey, group.contentKey)
                            }
                          >
                            {t('清空')}
                          </Button>
                        </div>
                      </div>
                      <Form.Input
                        field={group.subjectKey}
                        label={t('邮件标题')}
                      />
                      <Form.TextArea
                        field={group.contentKey}
                        label={t('邮件内容')}
                        autosize
                      />
                    </div>
                  ))}
                </div>
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
