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

import React from 'react';
import { Avatar, Typography, Table, Tag } from '@douyinfe/semi-ui';
import { IconCoinMoneyStroked } from '@douyinfe/semi-icons';
import {
  calculateModelPrice,
  getModelPriceItems,
  getPricingBillingColor,
  getPricingDisplayBillingLabel,
} from '../../../../../helpers';

const { Text } = Typography;

const soraSummaryWrapStyle = {
  border: '1px solid var(--semi-color-border)',
  borderRadius: 14,
  overflow: 'hidden',
  background: 'var(--semi-color-bg-0)',
  maxWidth: 320,
};

const soraSummaryGridStyle = {
  display: 'grid',
  gridTemplateColumns: 'minmax(0, 1fr) auto',
};

const soraSummaryHeaderCellStyle = {
  padding: '10px 14px',
  background: 'var(--semi-color-fill-0)',
  color: 'var(--semi-color-text-1)',
  fontSize: 13,
  fontWeight: 600,
};

const soraSummaryCellStyle = {
  padding: '10px 14px',
  borderTop: '1px solid var(--semi-color-border)',
  color: 'var(--semi-color-text-0)',
  fontSize: 13,
};

const renderSoraPriceSummary = (priceData, t) => {
  const tiers = Array.isArray(priceData?.resolutionTiers)
    ? priceData.resolutionTiers
    : [];

  if (tiers.length === 0) {
    return (
      <Text type='tertiary' size='small'>
        {t('\u6682\u65e0\u4ef7\u683c\u6863\u4f4d')}
      </Text>
    );
  }

  return (
    <div style={soraSummaryWrapStyle}>
      <div style={soraSummaryGridStyle}>
        <div style={soraSummaryHeaderCellStyle}>{t('分辨率')}</div>
        <div
          style={{
            ...soraSummaryHeaderCellStyle,
            textAlign: 'right',
          }}
        >
          {t('每秒价格')}
        </div>
      </div>
      {tiers.map((tier) => (
        <div
          key={tier.key || tier.label}
          style={soraSummaryGridStyle}
        >
          <div style={soraSummaryCellStyle}>{tier.label}</div>
          <div
            style={{
              ...soraSummaryCellStyle,
              textAlign: 'right',
              fontWeight: 600,
            }}
          >
            {`${tier.price}/s`}
          </div>
        </div>
      ))}
    </div>
  );
};

const ModelPricingTable = ({
  modelData,
  groupRatio,
  currency,
  siteDisplayType,
  tokenUnit,
  displayPrice,
  showRatio,
  usableGroup,
  autoGroups = [],
  t,
}) => {
  const modelEnableGroups = Array.isArray(modelData?.enable_groups)
    ? modelData.enable_groups
    : [];
  const autoChain = autoGroups.filter((group) =>
    modelEnableGroups.includes(group),
  );

  const renderGroupPriceTable = () => {
    const groupSource =
      modelEnableGroups.length > 0
        ? modelEnableGroups
        : Object.keys(usableGroup || {});
    const availableGroups = Array.from(
      new Set(groupSource.filter((group) => group && group !== 'auto')),
    );

    const tableData = availableGroups.map((group) => {
      const effectiveGroupRatio = {
        ...(groupRatio || {}),
        [group]: groupRatio?.[group] ?? 1,
      };
      const priceData = modelData
        ? calculateModelPrice({
            record: modelData,
            selectedGroup: group,
            groupRatio: effectiveGroupRatio,
            tokenUnit,
            displayPrice,
            currency,
            quotaDisplayType: siteDisplayType,
          })
        : { inputPrice: '-', outputPrice: '-', price: '-' };

      return {
        key: group,
        group,
        ratio: groupRatio?.[group] ?? priceData.usedGroupRatio ?? 1,
        billingType: getPricingDisplayBillingLabel(modelData, t),
        billingColor: getPricingBillingColor(modelData),
        priceItems: getModelPriceItems(priceData, t, siteDisplayType),
        priceData,
      };
    });

    const columns = [
      {
        title: t('\u5206\u7ec4'),
        dataIndex: 'group',
        render: (text) => (
          <Tag color='white' size='small' shape='circle'>
            {text}
            {t('\u5206\u7ec4')}
          </Tag>
        ),
      },
    ];

    const isDynamic = modelData?.billing_mode === 'tiered_expr';
    if (showRatio || isDynamic) {
      columns.push({
        title: t('\u5206\u7ec4\u500d\u7387'),
        dataIndex: 'ratio',
        render: (text) => (
          <Tag color='blue' size='small' shape='circle'>
            {text}x
          </Tag>
        ),
      });
    }

    columns.push({
      title: t('\u8ba1\u8d39\u7c7b\u578b'),
      dataIndex: 'billingType',
      render: (text, record) => (
        <Tag color={record.billingColor || 'white'} size='small' shape='circle'>
          {text || '-'}
        </Tag>
      ),
    });

    columns.push({
      title:
        siteDisplayType === 'TOKENS'
          ? t('\u8ba1\u8d39\u6458\u8981')
          : t('\u4ef7\u683c\u6458\u8981'),
      dataIndex: 'priceItems',
      render: (items, record) => {
        if (record.priceData?.isSoraParamPricing) {
          return renderSoraPriceSummary(record.priceData, t);
        }

        if (items.length === 1 && items[0].isDynamic) {
          return (
            <Text type='tertiary' size='small'>
              {t('\u89c1\u4e0a\u65b9\u52a8\u6001\u8ba1\u8d39\u8be6\u60c5')}
            </Text>
          );
        }

        return (
          <div className='space-y-1'>
            {items.map((item) => (
              <div key={item.key}>
                <div className='font-semibold text-orange-600'>
                  {item.label} {item.value}
                </div>
                <div className='text-xs text-gray-500'>{item.suffix}</div>
              </div>
            ))}
          </div>
        );
      },
    });

    return (
      <Table
        dataSource={tableData}
        columns={columns}
        pagination={false}
        size='small'
        bordered={false}
        className='!rounded-lg'
      />
    );
  };

  return (
    <div>
      <div className='flex items-center mb-4'>
        <Avatar size='small' color='orange' className='mr-2 shadow-md'>
          <IconCoinMoneyStroked size={16} />
        </Avatar>
        <div>
          <Text className='text-lg font-medium'>{t('\u5206\u7ec4\u4ef7\u683c')}</Text>
          <div className='text-xs text-gray-600'>
            {t('\u4e0d\u540c\u7528\u6237\u5206\u7ec4\u7684\u4ef7\u683c\u4fe1\u606f')}
          </div>
        </div>
      </div>

      {autoChain.length > 0 && (
        <div className='flex flex-wrap items-center gap-1 mb-4'>
          <span className='text-sm text-gray-600'>
            {t('auto\u5206\u7ec4\u8c03\u7528\u94fe\u8def')}
          </span>
          <span className='text-sm'>&rarr;</span>
          {autoChain.map((group, index) => (
            <React.Fragment key={group}>
              <Tag color='white' size='small' shape='circle'>
                {group}
                {t('\u5206\u7ec4')}
              </Tag>
              {index < autoChain.length - 1 && (
                <span className='text-sm'>&rarr;</span>
              )}
            </React.Fragment>
          ))}
        </div>
      )}

      {renderGroupPriceTable()}
    </div>
  );
};

export default ModelPricingTable;
