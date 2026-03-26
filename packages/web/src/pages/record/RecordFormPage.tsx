import { useEffect, useState, useCallback, useMemo } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import {
  Card,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Switch,
  Button,
  Space,
  message,
  Spin,
  AutoComplete,
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  useAuthStore,
  CURRENCIES,
  FUEL_GRADES,
  getFuelGradesByLocale,
  isElectricVehicle,
} from '@gastrack/shared';
import type { CreateFuelRecordRequest } from '@gastrack/shared';
import dayjs from 'dayjs';

/**
 * 自动计算逻辑：
 * - 加油量 × 单价 = 总费用
 * - 填写任意两个，自动算出第三个
 * - 只在另外两个字段都有值、且当前字段为空或未手动修改时才自动填充
 */

export default function RecordFormPage() {
  const { t, i18n: i18nInstance } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, recordId } = useParams<{
    vehicleId: string;
    recordId: string;
  }>();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [isEv, setIsEv] = useState(false);
  const [stationSuggestions, setStationSuggestions] = useState<string[]>([]);
  const [stationSearch, setStationSearch] = useState('');
  const user = useAuthStore((s) => s.user);

  const isEdit = !!recordId;

  // 根据用户设置推断默认单位
  const isImperial = user?.unit_system === 'imperial';
  const defaultFuelUnit = isImperial ? 'gal' : 'L';
  const defaultDistanceUnit = isImperial ? 'mi' : 'km';
  const defaultCurrency = user?.currency_code || 'CNY';

  // 加油站建议：根据输入文字模糊过滤
  const stationOptions = useMemo(() => {
    if (!stationSearch) {
      return stationSuggestions.map((name) => ({ value: name, label: name }));
    }
    const lower = stationSearch.toLowerCase();
    return stationSuggestions
      .filter((name) => name.toLowerCase().includes(lower))
      .map((name) => ({ value: name, label: name }));
  }, [stationSuggestions, stationSearch]);

  // 获取车辆信息以判断是否为电动车
  useEffect(() => {
    if (vehicleId) {
      vehicleApi.getById(vehicleId).then(({ data }) => {
        const vehicle = data.data;
        const ev = isElectricVehicle(vehicle.fuel_type);
        setIsEv(ev);
        // 电动车默认能量单位为 kWh
        if (ev && !isEdit) {
          form.setFieldValue('fuel_unit', 'kWh');
        }
        // 自动带入车辆的默认燃油标号（仅新建时）
        if (!ev && !isEdit && vehicle.fuel_grade) {
          form.setFieldValue('fuel_grade', vehicle.fuel_grade);
        }
      }).catch(() => { /* ignore */ });

      // 加载加油站/充电站建议列表
      fuelRecordApi.getStationSuggestions(vehicleId).then(({ data }) => {
        setStationSuggestions(data.data || []);
      }).catch(() => { /* ignore */ });
    }
  }, [vehicleId]);

  useEffect(() => {
    if (isEdit && vehicleId && recordId) {
      setFetching(true);
      fuelRecordApi
        .getById(vehicleId, recordId)
        .then(({ data }) => {
          const record = data.data;
          form.setFieldsValue({
            ...record,
            refuel_date: dayjs(record.refuel_date),
          });
        })
        .catch(() => message.error(t('common.error')))
        .finally(() => setFetching(false));
    }
  }, [vehicleId, recordId]);

  // 自动计算：根据已填写的任意两个值，算出第三个
  const autoCalc = useCallback(
    (changedField: 'fuel_amount' | 'unit_price' | 'total_cost') => {
      const amount = form.getFieldValue('fuel_amount') as number | undefined;
      const price = form.getFieldValue('unit_price') as number | undefined;
      const total = form.getFieldValue('total_cost') as number | undefined;

      if (changedField === 'fuel_amount' || changedField === 'unit_price') {
        // 修改了加油量或单价 → 算总费用
        if (amount && price) {
          form.setFieldValue('total_cost', Math.round(amount * price * 100) / 100);
        }
      }

      if (changedField === 'total_cost') {
        // 修改了总费用
        if (amount && total && !price) {
          // 有加油量无单价 → 算单价
          form.setFieldValue('unit_price', Math.round((total / amount) * 100) / 100);
        } else if (price && total && !amount) {
          // 有单价无加油量 → 算加油量
          form.setFieldValue('fuel_amount', Math.round((total / price) * 100) / 100);
        } else if (amount && total) {
          // 加油量和总费用都有，反推单价
          form.setFieldValue('unit_price', Math.round((total / amount) * 100) / 100);
        }
      }

      if (changedField === 'fuel_amount') {
        if (amount && total && !price) {
          form.setFieldValue('unit_price', Math.round((total / amount) * 100) / 100);
        }
      }

      if (changedField === 'unit_price') {
        if (price && total && !amount) {
          form.setFieldValue('fuel_amount', Math.round((total / price) * 100) / 100);
        }
      }
    },
    [form],
  );

  const onFinish = async (values: CreateFuelRecordRequest & { refuel_date: dayjs.Dayjs }) => {
    setLoading(true);
    const payload: CreateFuelRecordRequest = {
      ...values,
      refuel_date: values.refuel_date.format('YYYY-MM-DDTHH:mm:ssZ'),
      currency_code: values.currency_code || defaultCurrency,
      fuel_unit: values.fuel_unit || defaultFuelUnit,
      distance_unit: values.distance_unit || defaultDistanceUnit,
    };

    try {
      if (isEdit) {
        await fuelRecordApi.update(vehicleId!, recordId!, payload);
      } else {
        await fuelRecordApi.create(vehicleId!, payload);
      }
      message.success(t('common.success'));
      navigate(`/vehicles/${vehicleId}/records`);
    } catch (err: unknown) {
      const error = err as { response?: { data?: { message?: string } } };
      message.error(error.response?.data?.message || t('common.error'));
    } finally {
      setLoading(false);
    }
  };

  if (fetching) {
    return (
      <div style={{ textAlign: 'center', padding: 80 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/vehicles/${vehicleId}/records`)}
          />
          <h2>
            {isEdit ? t('fuelRecord.editRecord') : t('fuelRecord.addRecord')}
          </h2>
        </Space>
      </div>

      <Card style={{ maxWidth: 600 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={onFinish}
          initialValues={{
            refuel_date: dayjs(),
            is_full_tank: true,
            fuel_unit: defaultFuelUnit,
            currency_code: defaultCurrency,
            distance_unit: defaultDistanceUnit,
          }}
        >
          <Form.Item
            name="refuel_date"
            label={isEv ? t('fuelRecord.chargingDate') : t('fuelRecord.fuelDate')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <DatePicker
              showTime
              style={{ width: '100%' }}
              format="YYYY-MM-DD HH:mm"
            />
          </Form.Item>

          <Form.Item name="station_name" label={isEv ? t('fuelRecord.chargingStation') : t('fuelRecord.station')}>
            <AutoComplete
              options={stationOptions}
              onSearch={setStationSearch}
              placeholder={isEv ? t('fuelRecord.chargingStationPlaceholder') : t('fuelRecord.stationPlaceholder')}
              allowClear
              filterOption={false}
            />
          </Form.Item>

          {/* 加油量/充电量（单位跟随用户设置） */}
          <Form.Item
            name="fuel_amount"
            label={isEv ? t('fuelRecord.chargingAmount') : t('fuelRecord.fuelAmount')}
          >
            <InputNumber
              min={0.01}
              step={0.01}
              style={{ width: '100%' }}
              suffix={isEv ? 'kWh' : defaultFuelUnit}
              onChange={() => autoCalc('fuel_amount')}
            />
          </Form.Item>
          <Form.Item name="fuel_unit" hidden><Input /></Form.Item>

          {/* 单价（货币跟随用户设置） */}
          <Form.Item
            name="unit_price"
            label={t('fuelRecord.pricePerUnit')}
          >
            <InputNumber
              min={0.01}
              step={0.01}
              style={{ width: '100%' }}
              prefix={CURRENCIES.find((c) => c.value === defaultCurrency)?.symbol || defaultCurrency}
              suffix={`/${isEv ? 'kWh' : defaultFuelUnit}`}
              onChange={() => autoCalc('unit_price')}
            />
          </Form.Item>
          <Form.Item name="currency_code" hidden><Input /></Form.Item>

          {/* 总费用 */}
          <Form.Item
            name="total_cost"
            label={t('fuelRecord.totalCost')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <InputNumber
              min={0}
              step={0.01}
              style={{ width: '100%' }}
              prefix={CURRENCIES.find((c) => c.value === defaultCurrency)?.symbol || defaultCurrency}
              onChange={() => autoCalc('total_cost')}
            />
          </Form.Item>

          {/* 里程（单位跟随用户设置） */}
          <Form.Item
            name="odometer"
            label={t('fuelRecord.odometer')}
            rules={[{ required: true, message: t('common.required') }]}
          >
            <InputNumber min={0} step={1} style={{ width: '100%' }} suffix={defaultDistanceUnit} />
          </Form.Item>
          <Form.Item name="distance_unit" hidden><Input /></Form.Item>

          {/* 燃油标号（仅燃油车显示）——按当前语言优先显示本地区标号 */}
          {!isEv && (() => {
            const localGrades = getFuelGradesByLocale(i18nInstance.language);
            const allGrades = FUEL_GRADES;
            const localValues = new Set(localGrades.map((g) => g.value));
            const otherGrades = allGrades.filter((g) => !localValues.has(g.value));

            return (
              <Form.Item name="fuel_grade" label={t('fuelRecord.fuelGrade')}>
                <Select
                  allowClear
                  placeholder={t('fuelRecord.fuelGradePlaceholder')}
                  options={[
                    {
                      label: t('fuelGrade.localGroup'),
                      options: localGrades.map((g) => ({
                        value: g.value,
                        label: t(g.label),
                      })),
                    },
                    {
                      label: t('fuelGrade.otherGroup'),
                      options: otherGrades.map((g) => ({
                        value: g.value,
                        label: t(g.label),
                      })),
                    },
                  ]}
                />
              </Form.Item>
            );
          })()}

          <Form.Item
            name="is_full_tank"
            label={isEv ? t('fuelRecord.isFullCharge') : t('fuelRecord.isFullTank')}
            valuePropName="checked"
          >
            <Switch />
          </Form.Item>

          <Form.Item name="note" label={t('fuelRecord.notes')}>
            <Input.TextArea rows={3} placeholder={t('fuelRecord.notes')} />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {t('common.save')}
              </Button>
              <Button
                onClick={() => navigate(`/vehicles/${vehicleId}/records`)}
              >
                {t('common.cancel')}
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
