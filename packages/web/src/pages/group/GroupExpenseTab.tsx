import { useCallback, useEffect, useState } from 'react';
import { Card, Col, Empty, List, Row, Segmented, Select, Space, Statistic, Table, Typography } from 'antd';
import { CaretDownOutlined, CaretUpOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import { formatCurrency, formatNumber, groupApi } from '@gastrack/shared';
import type { GroupExpenseStatsResponse } from '@gastrack/shared';
import { useIsMobile } from '../../hooks/useIsMobile';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import UnconvertedNotice from '../../components/UnconvertedNotice';
import { useGroupUnits } from './useGroupUnits';
import Money from './Money';

const { Text } = Typography;

/** 群组费用看板 */
export default function GroupExpenseTab({ groupId }: { groupId: string }) {
  const { t } = useTranslation();
  const isMobile = useIsMobile();
  const {
    currency, efficiencyUnit, fuelUnit, distanceUnit,
    convertFuel, convertDistance, convertEfficiency, effDisplay,
  } = useGroupUnits();

  const [expenseStats, setExpenseStats] = useState<GroupExpenseStatsResponse | null>(null);
  const [expenseLoading, setExpenseLoading] = useState(true);
  const [expensePeriod, setExpensePeriod] = useState<'month' | 'year'>('month');
  const [expenseYear, setExpenseYear] = useState(new Date().getFullYear());
  const begin = useLatestRequest();

  const load = useCallback(async () => {
    const isLatest = begin();
    setExpenseLoading(true);
    try {
      const { data } = await groupApi.getExpenseStats(groupId, {
        period: expensePeriod,
        year: expensePeriod === 'month' ? expenseYear : undefined,
      });
      if (isLatest()) setExpenseStats(data.data);
    } catch {
      if (isLatest()) setExpenseStats(null);
    } finally {
      if (isLatest()) setExpenseLoading(false);
    }
  }, [groupId, expensePeriod, expenseYear, begin]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div>
      <Space wrap style={{ marginBottom: 16 }}>
        <Segmented
          value={expensePeriod}
          onChange={(val) => {
            const p = val as 'month' | 'year';
            setExpensePeriod(p);
          }}
          options={[
            { label: t('group.byMonth'), value: 'month' },
            { label: t('group.byYear'), value: 'year' },
          ]}
        />
        {expensePeriod === 'month' && (
          <Select
            value={expenseYear}
            onChange={(val) => {
              setExpenseYear(val);
            }}
            style={{ width: 100 }}
            options={Array.from({ length: 5 }, (_, i) => {
              const y = new Date().getFullYear() - i;
              return { label: `${y}`, value: y };
            })}
          />
        )}
      </Space>

      {expenseLoading ? (
        <Card loading />
      ) : expenseStats ? (
        <>
          <UnconvertedNotice currencies={[expenseStats.unconverted_currencies]} />
          {/* 统计卡片 */}
          <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
            {[
              {
                title: t('group.totalExpense'),
                value: { raw: 0, displayText: formatCurrency(expenseStats.summary.total_cost, expenseStats.currency_code || currency) },
                change: expenseStats.summary.cost_change_pct,
                precision: 2,
                isCost: true,
              },
              {
                title: t('group.totalFuelAmount'),
                value: { raw: convertFuel(expenseStats.summary.total_fuel), displayText: '' },
                change: expenseStats.summary.fuel_change_pct,
                precision: 2,
                suffix: fuelUnit,
                isCost: false,
              },
              {
                title: t('group.totalMileage'),
                value: { raw: convertDistance(expenseStats.summary.total_distance), displayText: '' },
                change: expenseStats.summary.distance_change_pct,
                precision: 0,
                suffix: distanceUnit,
                isCost: false,
              },
              {
                title: t('group.avgFuelEfficiency'),
                value: { raw: convertEfficiency(expenseStats.summary.avg_efficiency), displayText: '' },
                change: expenseStats.summary.efficiency_change_pct,
                precision: 2,
                suffix: efficiencyUnit,
                isCost: false,
              },
              // 有电动车数据时单独展示充电量与电耗（与燃油分开统计）
              ...((expenseStats.summary.total_energy ?? 0) > 0
                ? [
                    {
                      title: t('group.totalEnergyAmount'),
                      value: { raw: expenseStats.summary.total_energy ?? 0, displayText: '' },
                      change: 0,
                      precision: 2,
                      suffix: 'kWh',
                      isCost: false,
                    },
                    {
                      title: t('group.avgEnergyEfficiency'),
                      value: { raw: effDisplay(expenseStats.summary.avg_energy_efficiency ?? 0, true).value, displayText: '' },
                      change: 0,
                      precision: 2,
                      suffix: effDisplay(0, true).unit,
                      isCost: false,
                    },
                  ]
                : []),
            ].map((card, idx) => (
              <Col xs={12} sm={6} key={idx}>
                <Card size="small">
                  {card.isCost ? (
                    <Statistic
                      title={card.title}
                      formatter={() => (
                        <span>
                          {card.value.displayText}
                          {' '}
                          {card.change !== 0 && (
                            <Text
                              style={{ fontSize: 12 }}
                              type={card.change > 0 ? 'danger' : 'success'}
                            >
                              {card.change > 0 ? <CaretUpOutlined /> : <CaretDownOutlined />}
                              {formatNumber(Math.abs(card.change), 1)}%
                            </Text>
                          )}
                        </span>
                      )}
                    />
                  ) : (
                    <Statistic
                      title={card.title}
                      value={card.value.raw}
                      precision={card.precision}
                      suffix={
                        <span>
                          {card.suffix}{' '}
                          {card.change !== 0 && (
                            <Text
                              style={{ fontSize: 12 }}
                              type={card.change > 0 ? 'danger' : 'success'}
                            >
                              {card.change > 0 ? <CaretUpOutlined /> : <CaretDownOutlined />}
                              {formatNumber(Math.abs(card.change), 1)}%
                            </Text>
                          )}
                        </span>
                      }
                    />
                  )}
                </Card>
              </Col>
            ))}
          </Row>

          {/* 费用趋势表格 */}
          {expenseStats.trend_items.length > 0 && (
            <>
              <Text strong style={{ display: 'block', marginBottom: 12 }}>
                {t('group.costTrend')}
              </Text>
              <Table
                rowKey="period_label"
                size="small"
                pagination={false}
                scroll={isMobile ? { x: 500 } : undefined}
                dataSource={expenseStats.trend_items}
                columns={[
                  {
                    title: expensePeriod === 'month' ? t('stats.byMonth') : t('stats.byYear'),
                    dataIndex: 'period_label',
                    key: 'period_label',
                  },
                  {
                    title: t('group.totalExpense'),
                    dataIndex: 'total_cost',
                    key: 'total_cost',
                    render: (val: number) => formatCurrency(val, expenseStats.currency_code || currency),
                  },
                  {
                    title: t('group.totalFuelAmount'),
                    dataIndex: 'total_fuel',
                    key: 'total_fuel',
                    render: (val: number) => `${formatNumber(convertFuel(val), 2)} ${fuelUnit}`,
                  },
                  {
                    title: t('group.totalMileage'),
                    dataIndex: 'total_distance',
                    key: 'total_distance',
                    render: (val: number) => `${formatNumber(convertDistance(val), 0)} ${distanceUnit}`,
                  },
                  {
                    title: t('group.avgFuelEfficiency'),
                    dataIndex: 'avg_efficiency',
                    key: 'avg_efficiency',
                    render: (val: number) => `${formatNumber(convertEfficiency(val), 2)} ${efficiencyUnit}`,
                  },
                ]}
              />
            </>
          )}

          {/* 成员费用占比 */}
          {expenseStats.member_breakdown.length > 0 && (
            <>
              <Text strong style={{ display: 'block', marginTop: 24, marginBottom: 12 }}>
                {t('group.costBreakdown')}
              </Text>
                <List
                dataSource={expenseStats.member_breakdown}
                renderItem={(item) => (
                  <List.Item>
                    <List.Item.Meta
                      title={item.nickname}
                      description={
                        <span>
                          <Money amount={item.total_cost} currency={item.currency_code || currency} /> · {formatNumber(convertFuel(item.total_fuel), 2)} {fuelUnit}
                        </span>
                      }
                    />
                    <div style={{ minWidth: 60, textAlign: 'right' }}>
                      <Text strong>{formatNumber(item.percentage, 1)}%</Text>
                    </div>
                  </List.Item>
                )}
              />
            </>
          )}
        </>
      ) : (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('group.noExpenseData')}
        />
      )}
    </div>
  );
}
