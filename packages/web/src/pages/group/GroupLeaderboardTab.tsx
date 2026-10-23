import { useCallback, useEffect, useState } from 'react';
import { Avatar, Badge, Card, Empty, List, Segmented, Select, Space, Typography, theme } from 'antd';
import { useTranslation } from 'react-i18next';
import { formatNumber, groupApi } from '@gastrack/shared';
import type { LeaderboardEnergy, LeaderboardMetric, LeaderboardPeriod, LeaderboardResponse } from '@gastrack/shared';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import { useGroupUnits } from './useGroupUnits';
import Money from './Money';

const { Text } = Typography;

/** 群组排行榜 */
export default function GroupLeaderboardTab({ groupId }: { groupId: string }) {
  const { t } = useTranslation();
  const { currency, distanceUnit, convertDistance, effDisplay } = useGroupUnits();
  const { token } = theme.useToken();

  const [leaderboard, setLeaderboard] = useState<LeaderboardResponse | null>(null);
  const [leaderboardLoading, setLeaderboardLoading] = useState(true);
  const [lbMetric, setLbMetric] = useState<LeaderboardMetric>('efficiency');
  // 油耗排行只在同一能源类别内比较
  const [lbEnergy, setLbEnergy] = useState<LeaderboardEnergy>('fuel');
  const [lbPeriod, setLbPeriod] = useState<LeaderboardPeriod>('current_month');
  const begin = useLatestRequest();

  const load = useCallback(async () => {
    const isLatest = begin();
    setLeaderboardLoading(true);
    try {
      const { data } = await groupApi.getLeaderboard(groupId, { metric: lbMetric, period: lbPeriod, energy: lbEnergy });
      if (isLatest()) setLeaderboard(data.data);
    } catch {
      if (isLatest()) setLeaderboard(null);
    } finally {
      if (isLatest()) setLeaderboardLoading(false);
    }
  }, [groupId, lbMetric, lbPeriod, lbEnergy, begin]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div>
      <Space wrap style={{ marginBottom: 16 }}>
        <Segmented
          value={lbMetric}
          onChange={(val) => {
            const m = val as LeaderboardMetric;
            setLbMetric(m);
          }}
          options={[
            { label: t('group.metricEfficiency'), value: 'efficiency' },
            { label: t('group.metricCost'), value: 'cost' },
            { label: t('group.metricDistance'), value: 'distance' },
            { label: t('group.metricFrequency'), value: 'frequency' },
          ]}
        />
        {lbMetric === 'efficiency' && (
          <Segmented
            value={lbEnergy}
            onChange={(val) => {
              const e = val as LeaderboardEnergy;
              setLbEnergy(e);
            }}
            options={[
              { label: t('group.energyFuel'), value: 'fuel' },
              { label: t('group.energyElectric'), value: 'electric' },
            ]}
          />
        )}
        <Select
          value={lbPeriod}
          onChange={(val) => {
            setLbPeriod(val);
          }}
          style={{ width: 120 }}
          options={[
            { label: t('group.periodCurrentMonth'), value: 'current_month' },
            { label: t('group.periodLastMonth'), value: 'last_month' },
            { label: t('group.periodLast3Months'), value: 'last_3_months' },
            { label: t('group.periodCurrentYear'), value: 'current_year' },
          ]}
        />
      </Space>

      {leaderboard && leaderboard.rankings.length > 0 && (
        <div style={{ marginBottom: 12 }}>
          <Text type="secondary">
            {t('group.groupAvg')}:{' '}
            {lbMetric === 'cost' ? (
              <Text strong><Money amount={leaderboard.group_avg} currency={currency} /></Text>
            ) : (
              <>
                <Text strong>
                  {lbMetric === 'efficiency'
                    ? formatNumber(effDisplay(leaderboard.group_avg, leaderboard.unit === 'kWh/100km').value, 2)
                    : lbMetric === 'distance'
                    ? formatNumber(convertDistance(leaderboard.group_avg), 0)
                    : formatNumber(leaderboard.group_avg, 2)}
                </Text>{' '}
                {lbMetric === 'efficiency'
                  ? effDisplay(0, leaderboard.unit === 'kWh/100km').unit
                  : lbMetric === 'distance'
                  ? distanceUnit
                  : leaderboard.unit}
              </>
            )}
          </Text>
        </div>
      )}

      {leaderboardLoading ? (
        <Card loading />
      ) : leaderboard && leaderboard.rankings.length > 0 ? (
        <List
          dataSource={leaderboard.rankings}
          renderItem={(item) => {
            const medals = ['🥇', '🥈', '🥉'];
            const medal = item.rank <= 3 ? medals[item.rank - 1] : undefined;
            return (
              <List.Item
                style={{
                  background: item.is_self ? token.colorPrimaryBg : undefined,
                  borderRadius: 8,
                  padding: '12px 16px',
                  marginBottom: 4,
                }}
              >
                <List.Item.Meta
                  avatar={
                    medal ? (
                      <span style={{ fontSize: 24 }}>{medal}</span>
                    ) : (
                      <Avatar size="small" style={{ backgroundColor: token.colorFillSecondary, color: token.colorTextSecondary }}>
                        {item.rank}
                      </Avatar>
                    )
                  }
                  title={
                    <Space>
                      <Text strong={item.is_self}>
                        {item.nickname}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {item.vehicle_name}
                      </Text>
                      {item.is_self && <Badge count="✦" style={{ backgroundColor: token.colorPrimary }} />}
                    </Space>
                  }
                  description={
                    <Space size="large">
                      <Text>
                        {lbMetric === 'cost' ? (
                          <Money amount={item.value} currency={currency} />
                        ) : (
                          <>
                            {lbMetric === 'efficiency'
                              ? formatNumber(effDisplay(item.value, leaderboard.unit === 'kWh/100km').value, 2)
                              : lbMetric === 'distance'
                              ? formatNumber(convertDistance(item.value), 0)
                              : formatNumber(item.value, 2)}{' '}
                            {lbMetric === 'efficiency'
                              ? effDisplay(0, leaderboard.unit === 'kWh/100km').unit
                              : lbMetric === 'distance'
                              ? distanceUnit
                              : leaderboard.unit}
                          </>
                        )}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {t('group.recordCount')}: {item.record_count}
                      </Text>
                      <Text
                        type={item.diff_from_avg < 0 ? 'success' : item.diff_from_avg > 0 ? 'danger' : 'secondary'}
                        style={{ fontSize: 12 }}
                      >
                        {t('group.diffFromAvg')}: {item.diff_from_avg > 0 ? '+' : ''}{formatNumber(item.diff_from_avg, 1)}%
                      </Text>
                    </Space>
                  }
                />
              </List.Item>
            );
          }}
        />
      ) : (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('group.noLeaderboardData')}
        />
      )}
    </div>
  );
}
