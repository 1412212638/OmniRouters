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
import { Avatar, Table, Tag, Typography } from '@douyinfe/semi-ui';
import { IconPriceTag } from '@douyinfe/semi-icons';

const { Text } = Typography;

export default function SoraPricingBreakdown({ pricing, t }) {
  const tiers = Array.isArray(pricing?.resolution_tiers)
    ? pricing.resolution_tiers
        .map((tier, index) => ({
          key: `${tier?.value || 'tier'}-${index}`,
          resolution: tier?.value || '-',
          multiplier: Number(tier?.multiplier) || 0,
        }))
        .filter((tier) => tier.resolution !== '-' && tier.multiplier > 0)
    : [];

  if (!pricing?.enabled || tiers.length === 0) {
    return null;
  }

  const columns = [
    {
      title: t('档位'),
      dataIndex: 'resolution',
      render: (text) => (
        <Tag color='blue' size='small'>
          {text}
        </Tag>
      ),
    },
    {
      title: t('倍率'),
      dataIndex: 'multiplier',
      render: (value) => <Text strong>{value}x</Text>,
    },
  ];

  return (
    <div>
      <div className='flex items-center mb-4'>
        <Avatar size='small' color='orange' className='mr-2 shadow-md'>
          <IconPriceTag size={16} />
        </Avatar>
        <div>
          <Text className='text-lg font-medium'>{t('动态计费')}</Text>
          <div className='text-xs text-gray-600'>
            {t('当前模型按请求中的 seconds 与 resolution 档位动态计算价格。')}
          </div>
        </div>
      </div>

      <div
        style={{
          padding: '10px 12px',
          borderRadius: 8,
          background: 'var(--semi-color-fill-0)',
          marginBottom: 12,
        }}
      >
        <Text size='small'>
          {t('请求必须提供 resolution 和 seconds；最终价格 = 基础每秒单价 × seconds × resolution 倍率 × 分组倍率。')}
        </Text>
      </div>

      <Text strong className='text-sm' style={{ display: 'block', marginBottom: 8 }}>
        {t('分档价格表')}
      </Text>
      <Table
        dataSource={tiers}
        columns={columns}
        pagination={false}
        size='small'
        bordered={false}
        className='!rounded-lg'
      />
    </div>
  );
}
