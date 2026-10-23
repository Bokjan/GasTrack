import { useCallback, useEffect, useState } from 'react';
import { Card, Col, Empty, Row, Select, Space, Tag, Typography } from 'antd';
import { ArrowDownOutlined, ArrowUpOutlined, EnvironmentOutlined, MinusOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import dayjs from 'dayjs';
import { FUEL_GRADES, groupApi } from '@gastrack/shared';
import type { GroupStationStatsResponse } from '@gastrack/shared';
import { useCurrency } from '../../hooks/useCurrency';
import { useLatestRequest } from '../../hooks/useLatestRequest';
import { useGroupUnits } from './useGroupUnits';
import Money from './Money';

const { Text } = Typography;

/** 群组加油站推荐（单价保留加油站的原币种） */
export default function GroupStationTab({ groupId }: { groupId: string }) {
  const { t } = useTranslation();
  const { currency, stationPrice } = useGroupUnits();
  const { referenceText } = useCurrency();

  const [stationStats, setStationStats] = useState<GroupStationStatsResponse | null>(null);
  const [stationLoading, setStationLoading] = useState(true);
  const [stationFuelGrade, setStationFuelGrade] = useState('');
  const [stationMonths, setStationMonths] = useState(6);
  const [stationSortBy, setStationSortBy] = useState('frequency');
  const begin = useLatestRequest();

  const load = useCallback(async () => {
    const isLatest = begin();
    setStationLoading(true);
    try {
      const { data } = await groupApi.getStationStats(groupId, {
        fuel_grade: stationFuelGrade || undefined,
        months: stationMonths,
        sort_by: stationSortBy,
      });
      if (isLatest()) setStationStats(data.data);
    } catch {
      if (isLatest()) setStationStats(null);
    } finally {
      if (isLatest()) setStationLoading(false);
    }
  }, [groupId, stationFuelGrade, stationMonths, stationSortBy, begin]);

  useEffect(() => {
    load();
  }, [load]);

  return (
    <div>
      <Space wrap style={{ marginBottom: 16 }}>
        <Select
          value={stationFuelGrade || undefined}
          placeholder={t('group.fuelGradeFilter')}
          allowClear
          style={{ width: 140 }}
          onChange={(val) => {
            const v = val || '';
            setStationFuelGrade(v);
          }}
          options={FUEL_GRADES.map((item) => ({
            label: t(item.label),
            value: item.value,
          }))}
        />
        <Select
          value={stationMonths}
          style={{ width: 130 }}
          onChange={(val) => {
            setStationMonths(val);
          }}
          options={[
            { label: t('group.recentMonths', { count: 3 }), value: 3 },
            { label: t('group.recentMonths', { count: 6 }), value: 6 },
            { label: t('group.recentMonths', { count: 12 }), value: 12 },
          ]}
        />
        <Select
          value={stationSortBy}
          style={{ width: 120 }}
          onChange={(val) => {
            setStationSortBy(val);
          }}
          options={[
            { label: t('group.sortByFrequency'), value: 'frequency' },
            { label: t('group.sortByPrice'), value: 'price' },
            { label: t('group.sortByDate'), value: 'date' },
          ]}
        />
      </Space>

      {stationLoading ? (
        <Card loading />
      ) : stationStats && stationStats.stations.length > 0 ? (
        <Row gutter={[16, 16]}>
          {stationStats.stations.map((station, idx) => (
            <Col xs={24} sm={12} lg={8} key={idx}>
              <Card size="small" hoverable>
                <Space direction="vertical" style={{ width: '100%' }} size={8}>
                  <Space align="center">
                    <EnvironmentOutlined style={{ color: '#1677ff', fontSize: 16 }} />
                    <Text strong style={{ fontSize: 15 }}>{station.station_name}</Text>
                  </Space>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.avgPrice')}</Text>
                      <div>
                        <Text strong><Money amount={stationPrice(station.avg_unit_price, station.price_unit).value} currency={station.currency_code || currency} reference={referenceText(stationPrice(station.avg_unit_price, station.price_unit).value, station.currency_code || currency, currency)} /></Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>/{stationPrice(0, station.price_unit).unit}</Text>
                      </div>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.latestPrice')}</Text>
                      <div>
                        <Text strong><Money amount={stationPrice(station.latest_unit_price, station.price_unit).value} currency={station.currency_code || currency} reference={referenceText(stationPrice(station.latest_unit_price, station.price_unit).value, station.currency_code || currency, currency)} /></Text>
                        <Text type="secondary" style={{ fontSize: 12 }}>/{stationPrice(0, station.price_unit).unit} </Text>
                        <Text
                          type={station.price_trend === 'up' ? 'danger' : station.price_trend === 'down' ? 'success' : 'secondary'}
                          style={{ fontSize: 12 }}
                        >
                          {station.price_trend === 'up' && <><ArrowUpOutlined /> {t('group.priceTrendUp')}</>}
                          {station.price_trend === 'down' && <><ArrowDownOutlined /> {t('group.priceTrendDown')}</>}
                          {station.price_trend === 'stable' && <><MinusOutlined /> {t('group.priceTrendStable')}</>}
                        </Text>
                      </div>
                    </Col>
                  </Row>

                  <Row gutter={16}>
                    <Col span={12}>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.visitCount')}</Text>
                      <div><Text>{station.visit_count} {t('group.metricFrequency')}</Text></div>
                    </Col>
                    <Col span={12}>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.latestVisit')}</Text>
                      <div><Text>{dayjs(station.latest_visit).format('YYYY-MM-DD')}</Text></div>
                    </Col>
                  </Row>

                  {station.visitors.length > 0 && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.regulars')}</Text>
                      <div style={{ marginTop: 4 }}>
                        {station.visitors.map((v, vi) => (
                          <Tag key={vi} style={{ marginBottom: 4 }}>
                            {v.nickname} ({v.count})
                          </Tag>
                        ))}
                      </div>
                    </div>
                  )}

                  {station.fuel_grades_seen.length > 0 && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 12 }}>{t('group.fuelGrades')}</Text>
                      <div style={{ marginTop: 4 }}>
                        {station.fuel_grades_seen.map((g, gi) => {
                          const grade = FUEL_GRADES.find((fg) => fg.value === g);
                          return (
                            <Tag key={gi} color="blue" style={{ marginBottom: 4 }}>
                              {grade ? t(grade.label) : g}
                            </Tag>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </Space>
              </Card>
            </Col>
          ))}
        </Row>
      ) : (
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={t('group.noStationData')}
        />
      )}
    </div>
  );
}
