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
import {
  Card,
  Tag,
  Tooltip,
  Checkbox,
  Empty,
  Pagination,
  Button,
  Avatar,
} from '@douyinfe/semi-ui';
import { IconHelpCircle } from '@douyinfe/semi-icons';
import { Copy } from 'lucide-react';
import {
  IllustrationNoResult,
  IllustrationNoResultDark,
} from '@douyinfe/semi-illustrations';
import {
  stringToColor,
  calculateModelPrice,
  formatPriceInfo,
  formatDynamicPriceSummary,
  getPricingBillingColor,
  getPricingDisplayBillingLabel,
  getLobeHubIcon,
} from '../../../../../helpers';
import PricingCardSkeleton from './PricingCardSkeleton';
import { useMinimumLoadingTime } from '../../../../../hooks/common/useMinimumLoadingTime';
import { renderLimitedItems } from '../../../../common/ui/RenderUtils';
import { useIsMobile } from '../../../../../hooks/common/useIsMobile';

const CARD_STYLES = {
  container:
    'w-12 h-12 rounded-2xl flex items-center justify-center relative shadow-md',
  icon: 'w-8 h-8 flex items-center justify-center',
  selected: 'border-blue-500 bg-blue-50',
  default: 'border-gray-200 hover:border-gray-300',
};

const PricingCardView = ({
  filteredModels,
  loading,
  rowSelection,
  pageSize,
  setPageSize,
  currentPage,
  setCurrentPage,
  selectedGroup,
  groupRatio,
  copyText,
  setModalImageUrl,
  setIsModalOpenurl,
  currency,
  siteDisplayType,
  tokenUnit,
  displayPrice,
  showRatio,
  t,
  selectedRowKeys = [],
  setSelectedRowKeys,
  openModelDetail,
}) => {
  const showSkeleton = useMinimumLoadingTime(loading);
  const startIndex = (currentPage - 1) * pageSize;
  const paginatedModels = filteredModels.slice(
    startIndex,
    startIndex + pageSize,
  );
  const getModelKey = (model) => model.key ?? model.model_name ?? model.id;
  const isMobile = useIsMobile();

  const handleCheckboxChange = (model, checked) => {
    if (!setSelectedRowKeys) return;
    const modelKey = getModelKey(model);
    const newKeys = checked
      ? Array.from(new Set([...selectedRowKeys, modelKey]))
      : selectedRowKeys.filter((key) => key !== modelKey);
    setSelectedRowKeys(newKeys);
    rowSelection?.onChange?.(newKeys, null);
  };

  const getModelIcon = (model) => {
    if (!model || !model.model_name) {
      return (
        <div className={CARD_STYLES.container}>
          <Avatar size='large'>?</Avatar>
        </div>
      );
    }

    if (model.icon) {
      return (
        <div className={CARD_STYLES.container}>
          <div className={CARD_STYLES.icon}>
            {getLobeHubIcon(model.icon, 32)}
          </div>
        </div>
      );
    }

    if (model.vendor_icon) {
      return (
        <div className={CARD_STYLES.container}>
          <div className={CARD_STYLES.icon}>
            {getLobeHubIcon(model.vendor_icon, 32)}
          </div>
        </div>
      );
    }

    const avatarText = model.model_name.slice(0, 2).toUpperCase();
    return (
      <div className={CARD_STYLES.container}>
        <Avatar
          size='large'
          style={{
            width: 48,
            height: 48,
            borderRadius: 16,
            fontSize: 16,
            fontWeight: 'bold',
          }}
        >
          {avatarText}
        </Avatar>
      </div>
    );
  };

  const getModelDescription = (record) => record.description || '';

  const renderTags = (record) => {
    const billingTag = (
      <Tag
        key='billing'
        shape='circle'
        color={getPricingBillingColor(record)}
        size='small'
      >
        {getPricingDisplayBillingLabel(record, t)}
      </Tag>
    );

    const customTags = [];
    if (record.tags) {
      const tagArr = record.tags.split(',').filter(Boolean);
      tagArr.forEach((tg, idx) => {
        customTags.push(
          <Tag
            key={`custom-${idx}`}
            shape='circle'
            color={stringToColor(tg)}
            size='small'
          >
            {tg}
          </Tag>,
        );
      });
    }

    return (
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>{billingTag}</div>
        <div className='flex items-center gap-1'>
          {customTags.length > 0 &&
            renderLimitedItems({
              items: customTags.map((tag, idx) => ({
                key: `custom-${idx}`,
                element: tag,
              })),
              renderItem: (item) => item.element,
              maxDisplay: 3,
            })}
        </div>
      </div>
    );
  };

  const renderSoraSummaryTags = (priceData) => {
    const tiers = Array.isArray(priceData?.resolutionTiers)
      ? priceData.resolutionTiers
      : [];

    return (
      <div className='flex flex-wrap gap-2 pt-1'>
        <Tag color='orange' size='small' shape='circle'>
          {t('\u52a8\u6001\u8ba1\u8d39')}
        </Tag>
        {tiers.length > 0 && (
          <Tag color='grey' size='small' shape='circle'>
            {tiers.length}
            {t('\u6863')}
          </Tag>
        )}
        <Tag color='grey' size='small' shape='circle'>
          {t('\u6309\u79d2')}
        </Tag>
      </div>
    );
  };

  const renderCardPriceSummary = (priceData) => {
    if (priceData?.isSoraParamPricing) {
      return renderSoraSummaryTags(priceData);
    }

    return (
      <div className='flex flex-col gap-1 text-xs mt-2'>
        {priceData.isDynamicPricing ? (
          formatDynamicPriceSummary(
            priceData.billingExpr,
            t,
            priceData.usedGroupRatio,
          )
        ) : (
          formatPriceInfo(priceData, t, siteDisplayType)
        )}
      </div>
    );
  };

  if (showSkeleton) {
    return (
      <PricingCardSkeleton
        rowSelection={!!rowSelection}
        showRatio={showRatio}
      />
    );
  }

  if (!filteredModels || filteredModels.length === 0) {
    return (
      <div className='flex justify-center items-center py-20'>
        <Empty
          image={<IllustrationNoResult style={{ width: 150, height: 150 }} />}
          darkModeImage={
            <IllustrationNoResultDark style={{ width: 150, height: 150 }} />
          }
          description={t('\u641c\u7d22\u65e0\u7ed3\u679c')}
        />
      </div>
    );
  }

  return (
    <div className='px-2 pt-2'>
      <div className='grid grid-cols-1 xl:grid-cols-2 2xl:grid-cols-3 gap-4'>
        {paginatedModels.map((model, index) => {
          const modelKey = getModelKey(model);
          const isSelected = selectedRowKeys.includes(modelKey);
          const priceData = calculateModelPrice({
            record: model,
            selectedGroup,
            groupRatio,
            tokenUnit,
            displayPrice,
            currency,
            quotaDisplayType: siteDisplayType,
          });
          const description = getModelDescription(model);

          return (
            <Card
              key={modelKey || index}
              className={`!rounded-2xl transition-all duration-200 hover:shadow-lg border cursor-pointer ${
                isSelected ? CARD_STYLES.selected : CARD_STYLES.default
              }`}
              bodyStyle={{ height: '100%' }}
              onClick={() => openModelDetail && openModelDetail(model)}
            >
              <div className='flex flex-col h-full'>
                <div className='flex items-start justify-between mb-3'>
                  <div className='flex items-start space-x-3 flex-1 min-w-0'>
                    {getModelIcon(model)}
                    <div className='flex-1 min-w-0'>
                      <h3 className='text-lg font-bold text-gray-900 truncate'>
                        {model.model_name}
                      </h3>
                    </div>
                  </div>

                  <div className='flex items-center space-x-2 ml-3'>
                    <Button
                      size='small'
                      theme='outline'
                      type='tertiary'
                      icon={<Copy size={12} />}
                      onClick={(e) => {
                        e.stopPropagation();
                        copyText(model.model_name);
                      }}
                    />

                    {rowSelection && (
                      <Checkbox
                        checked={isSelected}
                        onChange={(e) => {
                          e.stopPropagation();
                          handleCheckboxChange(model, e.target.checked);
                        }}
                      />
                    )}
                  </div>
                </div>

                <div className='flex-1 flex flex-col gap-3 mb-4'>
                  {renderCardPriceSummary(priceData)}
                  {description && (
                    <p
                      className='text-xs line-clamp-2 leading-relaxed'
                      style={{
                        color: 'var(--semi-color-text-2)',
                        overflowWrap: 'anywhere',
                        wordBreak: 'break-word',
                      }}
                    >
                      {description}
                    </p>
                  )}
                </div>

                <div className='mt-auto'>
                  {renderTags(model)}

                  {showRatio && (
                    <div className='pt-3'>
                      <div className='flex items-center space-x-1 mb-2'>
                        <span className='text-xs font-medium text-gray-700'>
                          {t('\u500d\u7387\u4fe1\u606f')}
                        </span>
                        <Tooltip
                          content={t(
                            '\u500d\u7387\u662f\u4e3a\u4e86\u65b9\u4fbf\u6362\u7b97\u4e0d\u540c\u4ef7\u683c\u7684\u6a21\u578b',
                          )}
                        >
                          <IconHelpCircle
                            className='text-blue-500 cursor-pointer'
                            size='small'
                            onClick={(e) => {
                              e.stopPropagation();
                              setModalImageUrl('/ratio.png');
                              setIsModalOpenurl(true);
                            }}
                          />
                        </Tooltip>
                      </div>
                      <div className='grid grid-cols-3 gap-2 text-xs text-gray-600'>
                        <div>
                          {t('\u6a21\u578b')}:{' '}
                          {model.quota_type === 0 ? model.model_ratio : t('\u65e0')}
                        </div>
                        <div>
                          {t('\u8865\u5168')}:{' '}
                          {model.quota_type === 0
                            ? parseFloat(model.completion_ratio.toFixed(3))
                            : t('\u65e0')}
                        </div>
                        <div>
                          {t('\u5206\u7ec4')}: {priceData?.usedGroupRatio ?? '-'}
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </Card>
          );
        })}
      </div>

      {filteredModels.length > 0 && (
        <div className='flex justify-center mt-6 py-4 border-t pricing-pagination-divider'>
          <Pagination
            currentPage={currentPage}
            pageSize={pageSize}
            total={filteredModels.length}
            showSizeChanger={true}
            pageSizeOptions={[10, 20, 50, 100]}
            size={isMobile ? 'small' : 'default'}
            showQuickJumper={isMobile}
            onPageChange={(page) => setCurrentPage(page)}
            onPageSizeChange={(size) => {
              setPageSize(size);
              setCurrentPage(1);
            }}
          />
        </div>
      )}
    </div>
  );
};

export default PricingCardView;
