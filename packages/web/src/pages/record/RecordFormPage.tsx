import { useEffect, useState, useCallback } from 'react';
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
} from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { useTranslation } from 'react-i18next';
import {
  fuelRecordApi,
  vehicleApi,
  useAuthStore,
  CURRENCIES,
  FUEL_UNITS,
  ENERGY_UNITS,
  DISTANCE_UNITS,
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
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { vehicleId, recordId } = useParams<{
    vehicleId: string;
    recordId: string;
  }>();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [isEv, setIsEv] = useState(false);
  const user = useAuthStore((s) => s.user);

  const isEdit = !!recordId;

  // 获取车辆信息以判断是否为电动车
  useEffect(() => {
    if (vehicleId) {
      vehicleApi.getById(vehicleId).then(({ data }) => {
        const ev = isElectricVehicle(data.data.fuel_type);
        setIsEv(ev);
        // 电动车默认能量单位为 kWh
        if (ev && !isEdit) {
          form.setFieldValue('fuel_unit', 'kWh');
        }
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
      currency_code: values.currency_code || user?.currency_code || 'CNY',
      fuel_unit: values.fuel_unit || 'L',
      distance_unit: values.distance_unit || 'km',
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
            fuel_unit: 'L',
            currency_code: user?.currency_code || 'CNY',
            distance_unit: 'km',
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
            <Input placeholder={isEv ? t('fuelRecord.chargingStation') : t('fuelRecord.station')} />
          </Form.Item>

          {/* 加油量/充电量 + 燃油/能量单位 */}
          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="fuel_amount"
              label={isEv ? t('fuelRecord.chargingAmount') : t('fuelRecord.fuelAmount')}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={0.01}
                step={0.01}
                style={{ width: '100%' }}
                onChange={() => autoCalc('fuel_amount')}
              />
            </Form.Item>

            <Form.Item
              name="fuel_unit"
              label={isEv ? t('fuelRecord.energyUnit') : t('fuelRecord.fuelUnit')}
              style={{ width: 120 }}
            >
              <Select
                options={(isEv ? ENERGY_UNITS : FUEL_UNITS).map((u) => ({
                  value: u.value,
                  label: t(u.label),
                }))}
              />
            </Form.Item>
          </Space>

          {/* 单价 + 货币 */}
          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="unit_price"
              label={t('fuelRecord.pricePerUnit')}
              style={{ flex: 1 }}
            >
              <InputNumber
                min={0.01}
                step={0.01}
                style={{ width: '100%' }}
                onChange={() => autoCalc('unit_price')}
              />
            </Form.Item>

            <Form.Item
              name="currency_code"
              label={t('fuelRecord.currency')}
              style={{ width: 140 }}
            >
              <Select
                options={CURRENCIES.map((c) => ({
                  value: c.value,
                  label: c.label,
                }))}
              />
            </Form.Item>
          </Space>

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
              onChange={() => autoCalc('total_cost')}
            />
          </Form.Item>

          {/* 里程 + 里程单位 */}
          <Space style={{ width: '100%' }} size="middle">
            <Form.Item
              name="odometer"
              label={t('fuelRecord.odometer')}
              rules={[{ required: true, message: t('common.required') }]}
              style={{ flex: 1 }}
            >
              <InputNumber min={0} step={1} style={{ width: '100%' }} />
            </Form.Item>

            <Form.Item
              name="distance_unit"
              label={t('fuelRecord.distanceUnit')}
              style={{ width: 120 }}
            >
              <Select
                options={DISTANCE_UNITS.map((u) => ({
                  value: u.value,
                  label: t(u.label),
                }))}
              />
            </Form.Item>
          </Space>

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
